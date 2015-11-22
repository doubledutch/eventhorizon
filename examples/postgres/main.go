// +build postgres

package main

import (
	"os"

	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/examples/common"
)

func main() {
	eventBus := eventhorizon.NewInternalEventBus()
	conn := os.Getenv("POSTGRES_URL")

	newEventStore := func() (eventhorizon.RemoteEventStore, error) {
		return eventhorizon.NewPostgresEventStore(eventBus, conn)
	}

	newReadRepository := func(name string) (eventhorizon.RemoteReadRepository, error) {
		return eventhorizon.NewPostgresReadRepository(conn, name)
	}

	common.Run(eventBus, newEventStore, newReadRepository)
}
