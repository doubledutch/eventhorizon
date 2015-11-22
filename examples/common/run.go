package common

import (
	"fmt"
	"log"

	"github.com/looplab/eventhorizon"
)

type NewRemoteEventStoreFunc func() (eventhorizon.RemoteEventStore, error)

type NewReadRepositoryFunc func(string) (eventhorizon.RemoteReadRepository, error)

func Run(eventBus eventhorizon.EventBus, newEventStore NewRemoteEventStoreFunc,
	newReadRepository NewReadRepositoryFunc) {
	eventBus.AddGlobalHandler(&loggerSubscriber{})

	// Create the event store.
	eventStore, err := newEventStore()
	if err != nil {
		log.Fatalf("could not create event store: %s", err)
	}
	defer eventStore.Close()

	eventStore.RegisterEventType(&InviteCreated{}, func() eventhorizon.Event { return &InviteCreated{} })
	eventStore.RegisterEventType(&InviteAccepted{}, func() eventhorizon.Event { return &InviteAccepted{} })
	eventStore.RegisterEventType(&InviteDeclined{}, func() eventhorizon.Event { return &InviteDeclined{} })

	// Create the aggregate repository.
	repository, err := eventhorizon.NewCallbackRepository(eventStore)
	if err != nil {
		log.Fatalf("could not create repository: %s", err)
	}

	// Register an aggregate factory.
	repository.RegisterAggregate(&InvitationAggregate{},
		func(id eventhorizon.UUID) eventhorizon.Aggregate {
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
	commandBus := eventhorizon.NewInternalCommandBus()
	commandBus.SetHandler(handler, &CreateInvite{})
	commandBus.SetHandler(handler, &AcceptInvite{})
	commandBus.SetHandler(handler, &DeclineInvite{})

	// Create and register a read model for individual invitations.
	invitationRepository, err := newReadRepository("invitations")
	if err != nil {
		log.Fatalf("could not create invitation repository: %s", err)
	}
	defer invitationRepository.Close()
	invitationRepository.SetModel(func() interface{} { return &Invitation{} })
	invitationProjector := NewInvitationProjector(invitationRepository)
	eventBus.AddHandler(invitationProjector, &InviteCreated{})
	eventBus.AddHandler(invitationProjector, &InviteAccepted{})
	eventBus.AddHandler(invitationProjector, &InviteDeclined{})

	// Create and register a read model for a guest list.
	eventID := eventhorizon.NewUUID()
	guestListRepository, err := newReadRepository("guest_lists")
	if err != nil {
		log.Fatalf("could not create guest list repository: %s", err)
	}
	defer guestListRepository.Close()
	guestListRepository.SetModel(func() interface{} { return &GuestList{} })
	guestListProjector := NewGuestListProjector(guestListRepository, eventID)
	eventBus.AddHandler(guestListProjector, &InviteCreated{})
	eventBus.AddHandler(guestListProjector, &InviteAccepted{})
	eventBus.AddHandler(guestListProjector, &InviteDeclined{})

	// Clear DB collections.
	eventStore.Clear()
	invitationRepository.Clear()
	guestListRepository.Clear()

	// Issue some invitations and responses.
	// Note that Athena tries to decline the event, but that is not allowed
	// by the domain logic in InvitationAggregate. The result is that she is
	// still accepted.
	athenaID := eventhorizon.NewUUID()
	commandBus.HandleCommand(&CreateInvite{InvitationID: athenaID, EventID: eventID, Name: "Athena", Age: 42})
	commandBus.HandleCommand(&AcceptInvite{InvitationID: athenaID})
	err = commandBus.HandleCommand(&DeclineInvite{InvitationID: athenaID})
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	hadesID := eventhorizon.NewUUID()
	commandBus.HandleCommand(&CreateInvite{InvitationID: hadesID, EventID: eventID, Name: "Hades"})
	commandBus.HandleCommand(&AcceptInvite{InvitationID: hadesID})

	zeusID := eventhorizon.NewUUID()
	commandBus.HandleCommand(&CreateInvite{InvitationID: zeusID, EventID: eventID, Name: "Zeus"})
	commandBus.HandleCommand(&DeclineInvite{InvitationID: zeusID})

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
