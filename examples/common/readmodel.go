// Copyright (c) 2014 - Max Persson <max@looplab.se>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"log"

	"github.com/looplab/eventhorizon"
)

type Invitation struct {
	ID     eventhorizon.UUID `json:"id"`
	Name   string            `json:"name"`
	Status string            `json:"status"`
}

// Projector that writes to a read model

type InvitationProjector struct {
	repository eventhorizon.ReadRepository
}

func NewInvitationProjector(repository eventhorizon.ReadRepository) *InvitationProjector {
	p := &InvitationProjector{
		repository: repository,
	}
	return p
}

func (p *InvitationProjector) HandleEvent(event eventhorizon.Event) {
	switch event := event.(type) {
	case *InviteCreated:
		i := &Invitation{
			ID:   event.InvitationID,
			Name: event.Name,
		}
		if err := p.repository.Save(i.ID, i); err != nil {
			log.Fatalf("Unable to save event for invitation created: %s", err)
		}
	case *InviteAccepted:
		m, err := p.repository.Find(event.InvitationID)
		if err != nil {
			log.Fatalf("Unable to find model for invite accepted: %s", err)
		}
		i := m.(*Invitation)
		i.Status = "accepted"
		if err := p.repository.Save(i.ID, i); err != nil {
			log.Fatalf("Unable to save invite accepted event: %s", err)
		}
	case *InviteDeclined:
		m, err := p.repository.Find(event.InvitationID)
		if err != nil {
			log.Fatalf("Unable to find model for invite declined: %s", err)
		}
		i := m.(*Invitation)
		i.Status = "declined"
		if err := p.repository.Save(i.ID, i); err != nil {
			log.Fatalf("Unable to save invite declined event: %s", err)
		}
	}
}

type GuestList struct {
	ID          eventhorizon.UUID `json:"id"`
	NumGuests   int               `json:"num_guests"`
	NumAccepted int               `json:"num_accepted"`
	NumDeclined int               `json:"num_declined"`
}

// Projector that writes to a read model

// GuestListProjector projects guest lists. Note, it currently only works for
// one guest list.
type GuestListProjector struct {
	repository eventhorizon.ReadRepository
	eventID    eventhorizon.UUID
}

// NewGuestListProjector creates a new GuestListProjector.
func NewGuestListProjector(repository eventhorizon.ReadRepository, eventID eventhorizon.UUID) *GuestListProjector {
	p := &GuestListProjector{
		repository: repository,
		eventID:    eventID,
	}
	return p
}

// HandleEvent handles events that GuestListProjector is concerned about.
// HandleEvent builds guest lists.
func (p *GuestListProjector) HandleEvent(event eventhorizon.Event) {
	switch event.(type) {
	case *InviteCreated:
		m, err := p.repository.Find(p.eventID)
		if err != nil && err != eventhorizon.ErrModelNotFound {
			log.Fatalf("guest list: unable to find model for invite created: %s", err)
		}
		if m == nil {
			m = &GuestList{
				ID: p.eventID,
			}
		}
		g := m.(*GuestList)
		if err := p.repository.Save(p.eventID, g); err != nil {
			log.Fatalf("guest list: unable to save event: %s", err)
		}
	case *InviteAccepted:
		m, err := p.repository.Find(p.eventID)
		if err != nil {
			log.Fatalf("guest list: unable to find model for invite accepted: %s", err)
		}
		g := m.(*GuestList)
		g.NumAccepted++
		if err := p.repository.Save(p.eventID, g); err != nil {
			log.Fatalf("guest list: unable to save event: %s", err)
		}
	case *InviteDeclined:
		m, err := p.repository.Find(p.eventID)
		if err != nil {
			log.Fatalf("guest list: unable to find model for invite declined: %s", err)
		}
		g := m.(*GuestList)
		g.NumDeclined++
		if err := p.repository.Save(p.eventID, g); err != nil {
			log.Fatalf("guest list: unable to save event: %s", err)
		}
	}
}
