package eventhorizon

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/doubledutch/lager"
	"github.com/streadway/amqp"
)

const (
	eventExchange     = "events.exchange"
	eventExchangeType = "topic"
	eventQueueName    = "events.queue"
	eventKey          = "#"
)

func appExchange(app, exchange string) string {
	return strings.Join([]string{app, exchange}, ".")
}

func appTagQueueName(app, tag, queueName string) string {
	return strings.Join([]string{app, tag, queueName}, ".")
}

// RabbitMQEventBus implements CommandBus using RabbitMQ.
type RabbitMQEventBus struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	eventHandlersLock sync.Mutex
	eventHandlers     map[string]map[EventHandler]bool
	factoriesLock     sync.Mutex
	factories         map[string]func() Event

	localHandlers  map[EventHandler]bool
	globalHandlers map[EventHandler]bool

	done chan error

	application string
	exchange    string
	queue       string
	tag         string

	lgr lager.ContextLager
}

// NewRabbitMQEventBus creates a new RabbitMQ event bus. amqpURI is the RabbitMQ
// URI for rabbitmq. app is provides a namespace for this application, allowing
// for multiple event buses to run on one RabbitMQ and not conflict with eachother.
// tag is used as the RabbitMQ consumer tag for this bus.
func NewRabbitMQEventBus(amqpURI, app, tag string) (*RabbitMQEventBus, error) {
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

	exchange := appExchange(app, eventExchange)
	queueName := appTagQueueName(app, tag, eventQueueName)

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

	bus := &RabbitMQEventBus{
		conn:           connection,
		channel:        channel,
		exchange:       exchange,
		queue:          queueName,
		tag:            tag,
		eventHandlers:  make(map[string]map[EventHandler]bool),
		localHandlers:  make(map[EventHandler]bool),
		globalHandlers: make(map[EventHandler]bool),
		factories:      make(map[string]func() Event),
		done:           make(chan error),
		lgr:            lgr,
	}

	go bus.handleEvents(deliveries, bus.done)
	return bus, nil
}

// PublishEvent publishes a command to the commands exchange.
func (b *RabbitMQEventBus) PublishEvent(event Event) {
	// Send it locally
	for handler := range b.localHandlers {
		handler.HandleEvent(event)
	}

	// Send it to the queue
	d, err := json.Marshal(event)
	if err != nil {
		b.lgr.WithError(err).With(map[string]string{
			"event": event.AggregateID(),
		}).Errorf("Unable to publish event")
		return
	}

	b.channel.Publish(
		b.exchange,        // publish to an exchange
		event.EventType(), // routing to 0 or more queues
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            d,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		})
}

// Close closes the command bus, closing the rabbitmq connection.
func (b *RabbitMQEventBus) Close() error {
	// will close() the deliveries channel
	if err := b.channel.Cancel(b.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	return <-b.done
}

func (b *RabbitMQEventBus) handleEvents(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		b.factoriesLock.Lock()
		f, ok := b.factories[d.RoutingKey]
		b.factoriesLock.Unlock()
		if !ok {
			d.Reject(false)
			continue
		}

		event := f()
		if err := json.Unmarshal(d.Body, event); err != nil {
			b.lgr.WithError(err).With(map[string]string{
				"eventType": d.RoutingKey,
			}).Errorf("Unable to unmarshal received event")
			d.Reject(false)
			continue
		}

		if err := b.handleEvent(event); err != nil {
			d.Reject(false)
			continue
		}

		d.Ack(false)
	}
	done <- nil
}

func (b *RabbitMQEventBus) handleEvent(event Event) error {
	b.eventHandlersLock.Lock()
	handlers, ok := b.eventHandlers[event.EventType()]
	b.eventHandlersLock.Unlock()
	if ok {
		for handler := range handlers {
			handler.HandleEvent(event)
		}
	}

	for handler := range b.globalHandlers {
		handler.HandleEvent(event)
	}

	return nil
}

// RegisterEventType registers a event factory for a specific event.
func (b *RabbitMQEventBus) RegisterEventType(event Event, factory func() Event) error {
	b.factoriesLock.Lock()
	defer b.factoriesLock.Unlock()
	if _, ok := b.factories[event.EventType()]; ok {
		return ErrHandlerAlreadySet
	}

	b.factories[event.EventType()] = factory

	return nil
}

// AddHandler adds a handler for a specific local event.
func (b *RabbitMQEventBus) AddHandler(handler EventHandler, event Event) {
	b.eventHandlersLock.Lock()
	defer b.eventHandlersLock.Unlock()
	if _, ok := b.eventHandlers[event.EventType()]; !ok {
		b.eventHandlers[event.EventType()] = make(map[EventHandler]bool)
	}

	// Add handler to event type.
	b.eventHandlers[event.EventType()][handler] = true
}

// AddLocalHandler adds a handler for local events.
func (b *RabbitMQEventBus) AddLocalHandler(handler EventHandler) {
	b.localHandlers[handler] = true
}

// AddGlobalHandler adds a handler for global (remote) events.
func (b *RabbitMQEventBus) AddGlobalHandler(handler EventHandler) {
	b.globalHandlers[handler] = true
}
