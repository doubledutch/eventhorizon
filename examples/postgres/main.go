package main

import (
	"log"
	"os"

	"github.com/doubledutch/lager"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/examples/common"
)

func main() {
	lager.SetLevels(lager.LevelsFromString(os.Getenv("LOG_LEVELS")))

	var eventBus eventhorizon.EventBus
	var commandBus eventhorizon.CommandBus
	if uri := os.Getenv("RABBITMQ_URI"); uri != "" {
		remoteEventBus, err := eventhorizon.NewRabbitMQEventBus(uri, "test", "test")
		if err != nil {
			log.Fatalln("Unable to create rabbitmq event bus:", err)
		}
		defer remoteEventBus.Close()
		eventBus = remoteEventBus

		remoteCommandBus, err := eventhorizon.NewRabbitMQCommandBus(uri, "test", "test")
		if err != nil {
			log.Fatalln("Unable to create rabbitmq command bus:", err)
		}
		defer remoteEventBus.Close()
		commandBus = remoteCommandBus
	} else {
		eventBus = eventhorizon.NewInternalEventBus()
		commandBus = eventhorizon.NewInternalCommandBus()
	}

	conn := os.Getenv("POSTGRES_URL")

	newEventStore := func() (eventhorizon.EventStore, error) {
		return eventhorizon.NewPostgresEventStore(eventBus, conn)
	}

	newReadRepository := func(name string) (eventhorizon.ReadRepository, error) {
		return eventhorizon.NewPostgresReadRepository(conn, name)
	}

	common.Run(eventBus, commandBus, newEventStore, newReadRepository)
}
