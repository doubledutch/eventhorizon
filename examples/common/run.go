package common

import (
	"fmt"
	"log"
	"time"

	"github.com/looplab/eventhorizon"
	"github.com/odeke-em/go-uuid"
)

// NewEventStoreFunc creates a new EventStore
type NewEventStoreFunc func() (eventhorizon.EventStore, error)

// NewReadRepositoryFunc creates a new ReadRepository
type NewReadRepositoryFunc func(string) (eventhorizon.ReadRepository, error)

// Run runs the test scenario with the given EventStore and CommandBus.
// EventStores and ReadRepositories are created as needed with the given functions.
func Run(eventBus eventhorizon.EventBus, commandBus eventhorizon.CommandBus,
	newEventStore NewEventStoreFunc,
	newReadRepository NewReadRepositoryFunc) {
	eventBus.AddGlobalHandler(&loggerSubscriber{})

	if remoteEventBus, ok := eventBus.(eventhorizon.RemoteEventBus); ok {
		fmt.Println("event registered")
		remoteEventBus.RegisterEventType(&InviteCreated{}, func() eventhorizon.Event { return &InviteCreated{} })
		remoteEventBus.RegisterEventType(&InviteAccepted{}, func() eventhorizon.Event { return &InviteAccepted{} })
		remoteEventBus.RegisterEventType(&InviteDeclined{}, func() eventhorizon.Event { return &InviteDeclined{} })
		defer remoteEventBus.Close()
	}

	if remoteCommandBus, ok := commandBus.(eventhorizon.RemoteCommandBus); ok {
		fmt.Println("command registered")
		remoteCommandBus.RegisterCommandType(&CreateInvite{}, func() eventhorizon.Command { return &CreateInvite{} })
		remoteCommandBus.RegisterCommandType(&AcceptInvite{}, func() eventhorizon.Command { return &AcceptInvite{} })
		remoteCommandBus.RegisterCommandType(&DeclineInvite{}, func() eventhorizon.Command { return &DeclineInvite{} })
		defer remoteCommandBus.Close()
	}

	// Create the event store.
	eventStore, err := newEventStore()
	if err != nil {
		log.Fatalf("could not create event store: %s", err)
	}
	if remoteEventStore, ok := eventStore.(eventhorizon.RemoteEventStore); ok {
		remoteEventStore.RegisterEventType(&InviteCreated{}, func() eventhorizon.Event { return &InviteCreated{} })
		remoteEventStore.RegisterEventType(&InviteAccepted{}, func() eventhorizon.Event { return &InviteAccepted{} })
		remoteEventStore.RegisterEventType(&InviteDeclined{}, func() eventhorizon.Event { return &InviteDeclined{} })
		remoteEventStore.Clear()
		defer remoteEventStore.Close()
	}

	// Create the aggregate repository.
	repository, err := eventhorizon.NewCallbackRepository(eventStore)
	if err != nil {
		log.Fatalf("could not create repository: %s", err)
	}

	// Register an aggregate factory.
	repository.RegisterAggregate(&InvitationAggregate{},
		func(id string) eventhorizon.Aggregate {
			return &InvitationAggregate{
				AggregateBase: eventhorizon.NewAggregateBase(id),
			}
		},
	)

	// Create the aggregate command handler.
	handler, err := eventhorizon.NewAggregateCommandHandler(repository)
	if err != nil {
		log.Fatalf("could not create command handler: %s", err)
	}

	// Register the domain aggregates with the dispather. Remember to check for
	// errors here in a real app!
	handler.SetAggregate(&InvitationAggregate{}, &CreateInvite{})
	handler.SetAggregate(&InvitationAggregate{}, &AcceptInvite{})
	handler.SetAggregate(&InvitationAggregate{}, &DeclineInvite{})

	// Create the command bus and register the handler for the commands.
	commandBus.SetHandler(handler, &CreateInvite{})
	commandBus.SetHandler(handler, &AcceptInvite{})
	commandBus.SetHandler(handler, &DeclineInvite{})

	// Create and register a read model for individual invitations.
	invitationRepository, err := newReadRepository("invitations")
	if err != nil {
		log.Fatalf("could not create invitation repository: %s", err)
	}
	if remoteRepository, ok := invitationRepository.(eventhorizon.RemoteReadRepository); ok {
		remoteRepository.SetModel(func() interface{} { return &Invitation{} })
		remoteRepository.Clear()
		defer remoteRepository.Close()
	}
	invitationProjector := NewInvitationProjector(invitationRepository)
	eventBus.AddHandler(invitationProjector, &InviteCreated{})
	eventBus.AddHandler(invitationProjector, &InviteAccepted{})
	eventBus.AddHandler(invitationProjector, &InviteDeclined{})

	// Create and register a read model for a guest list.
	eventID := uuid.New()
	guestListRepository, err := newReadRepository("guest_lists")
	if err != nil {
		log.Fatalf("could not create guest list repository: %s", err)
	}
	if remoteRepository, ok := guestListRepository.(eventhorizon.RemoteReadRepository); ok {
		remoteRepository.SetModel(func() interface{} { return &GuestList{} })
		remoteRepository.Clear()
		defer remoteRepository.Close()
	}
	guestListProjector := NewGuestListProjector(guestListRepository, eventID)
	eventBus.AddHandler(guestListProjector, &InviteCreated{})
	eventBus.AddHandler(guestListProjector, &InviteAccepted{})
	eventBus.AddHandler(guestListProjector, &InviteDeclined{})

	// Issue some invitations and responses.
	// Note that Athena tries to decline the event, but that is not allowed
	// by the domain logic in InvitationAggregate. The result is that she is
	// still accepted.
	athenaID := uuid.New()
	commandBus.PublishCommand(&CreateInvite{InvitationID: athenaID, EventID: eventID, Name: "Athena", Age: 42})
	commandBus.PublishCommand(&AcceptInvite{InvitationID: athenaID})
	err = commandBus.PublishCommand(&DeclineInvite{InvitationID: athenaID})
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	hadesID := uuid.New()
	commandBus.PublishCommand(&CreateInvite{InvitationID: hadesID, EventID: eventID, Name: "Hades"})
	commandBus.PublishCommand(&AcceptInvite{InvitationID: hadesID})

	zeusID := uuid.New()
	commandBus.PublishCommand(&CreateInvite{InvitationID: zeusID, EventID: eventID, Name: "Zeus"})
	commandBus.PublishCommand(&DeclineInvite{InvitationID: zeusID})

	// TODO: Find a better way to ensure that all events were processed
	time.Sleep(1 * time.Second)

	// Read all invites.
	invitations, _ := invitationRepository.FindAll()
	for _, i := range invitations {
		fmt.Printf("invitation: %#v\n", i)
	}

	// Read the guest list.
	guestList, _ := guestListRepository.Find(eventID)
	fmt.Printf("guest list: %#v\n", guestList)
}

type loggerSubscriber struct{}

func (l *loggerSubscriber) HandleEvent(event eventhorizon.Event) {
	log.Printf("event: %#v\n", event)
}
