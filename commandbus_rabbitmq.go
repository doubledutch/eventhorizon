package eventhorizon

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/doubledutch/lager"
	"github.com/streadway/amqp"
)

const (
	commandExchange     = "commands.exchange"
	commandExchangeType = "topic"
	commandQueueName    = "commands.queue"
	commandKey          = "#"
)

// RabbitMQCommandBus implements CommandBus using RabbitMQ.
type RabbitMQCommandBus struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	handlersLock  sync.Mutex
	handlers      map[string]CommandHandler
	factoriesLock sync.Mutex
	factories     map[string]func() Command
	done          chan error

	exchange string
	queue    string
	tag      string

	lgr lager.ContextLager
}

// NewRabbitMQCommandBus creates a new RabbitMQ command bus. amqpURI is the RabbitMQ
// URI for rabbitmq. app is provides a namespace for this application, allowing
// for multiple command buses to run on one RabbitMQ and not conflict with eachother.
// tag is used as the RabbitMQ consumer tag for this bus.
func NewRabbitMQCommandBus(amqpURI, app, tag string) (*RabbitMQCommandBus, error) {
	lgr := lager.Child()

	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		lgr.WithError(err).Errorf("Error dialing URI: %s", amqpURI)
		return nil, fmt.Errorf("Dial err: %s", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		lgr.WithError(err).Errorf("Error opening channel")
		return nil, fmt.Errorf("Channel: %s", err)
	}

	exchange := appExchange(app, commandExchange)
	queueName := appTagQueueName(app, tag, commandQueueName)

	if err := channel.ExchangeDeclare(
		exchange,            // name
		commandExchangeType, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // noWait
		nil,                 // arguments
	); err != nil {
		channel.Close()
		connection.Close()
		lgr.WithError(err).Errorf("Error declaring exchange :%s", exchange)
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	queue, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		connection.Close()
		lgr.WithError(err).Errorf("Error declaring queue %s", queueName)
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	if err = channel.QueueBind(
		queue.Name, // name of the queue
		commandKey, // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		channel.Close()
		connection.Close()
		lgr.WithError(err).Errorf("Error binding queue %s", queueName)
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	deliveries, err := channel.Consume(
		queue.Name, // name
		tag,        // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		channel.Close()
		connection.Close()
		lgr.WithError(err).Errorf("Error consuming queue %s", queueName)
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	bus := &RabbitMQCommandBus{
		conn:      connection,
		channel:   channel,
		exchange:  exchange,
		queue:     queueName,
		tag:       tag,
		handlers:  make(map[string]CommandHandler),
		factories: make(map[string]func() Command),
		done:      make(chan error),
		lgr:       lgr,
	}

	go bus.handleCommands(deliveries, bus.done)
	return bus, nil
}

// PublishCommand publishes a command to the commands exchange.
func (b *RabbitMQCommandBus) PublishCommand(command Command) error {
	d, err := json.Marshal(command)
	if err != nil {
		b.lgr.WithError(err).Errorf("Unable to marshal command")
		return err
	}

	err = b.channel.Publish(
		b.exchange,            // publish to an exchange
		command.CommandType(), // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            d,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		})

	if err != nil {
		b.lgr.WithError(err).Errorf("Unable to publish command")
	}

	return err
}

// Close closes the command bus, closing the rabbitmq connection.
func (b *RabbitMQCommandBus) Close() error {
	// will close() the deliveries channel
	if err := b.channel.Cancel(b.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	return <-b.done
}

func (b *RabbitMQCommandBus) handleCommands(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		b.factoriesLock.Lock()
		f, ok := b.factories[d.RoutingKey]
		b.factoriesLock.Unlock()
		if !ok {
			b.lgr.With(map[string]string{
				"commandType": d.RoutingKey,
			}).Debugf("No factory for command type")
			d.Reject(false)
			continue
		}

		command := f()
		if err := json.Unmarshal(d.Body, command); err != nil {
			b.lgr.WithError(err).With(map[string]string{
				"commandType": d.RoutingKey,
			}).Errorf("Unable to unmarshal received command")
			d.Reject(false)
			continue
		}

		if err := b.HandleCommand(command); err != nil {
			b.lgr.WithError(err).With(map[string]string{
				"commandType": d.RoutingKey,
			}).Errorf("Error handling command")
			d.Reject(false)
			continue
		}

		b.lgr.With(map[string]string{"commandType": d.RoutingKey}).Debugf("Handled command")
		d.Ack(false)
	}
	done <- nil
}

// HandleCommand handles a command, dispatching it to the proper handlers.
func (b *RabbitMQCommandBus) HandleCommand(command Command) error {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()
	if handler, ok := b.handlers[command.CommandType()]; ok {
		return handler.HandleCommand(command)
	}
	return ErrHandlerNotFound
}

// SetHandler sets a handler for a specific command.
func (b *RabbitMQCommandBus) SetHandler(handler CommandHandler, command Command) error {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()
	if _, ok := b.handlers[command.CommandType()]; ok {
		return ErrHandlerAlreadySet
	}
	b.handlers[command.CommandType()] = handler
	return nil
}

// RegisterCommandType registers a command factory for a specific command.
func (b *RabbitMQCommandBus) RegisterCommandType(command Command, factory func() Command) error {
	b.factoriesLock.Lock()
	defer b.factoriesLock.Unlock()
	if _, ok := b.factories[command.CommandType()]; ok {
		return ErrHandlerAlreadySet
	}

	b.factories[command.CommandType()] = factory

	return nil
}
