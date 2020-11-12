package coordinator

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
)

type Coordinator struct {
	directory map[uuid.UUID]utils.Messenger
	messenger utils.Messenger
}

func (c *Coordinator) GetMessenger(id uuid.UUID) utils.Messenger {
	return c.directory[id]
}

func (c *Coordinator) NewMessenger(initialMessage utils.Message) utils.Messenger {
	id := uuid.New()

	c.directory[id] = utils.Messenger{
		Me:      id,
		Inbox:   make(chan utils.Message, 100),
		Outbox:  c.messenger.Inbox,
		Current: initialMessage,
	}

	return c.directory[id]
}

func InitCoordinator(coordinatorOutput chan utils.Message) (Coordinator, utils.Messenger) {
	coordinator := Coordinator{
		directory: make(map[uuid.UUID]utils.Messenger),
		messenger: utils.Messenger{
			Me:     uuid.New(),
			Inbox:  make(chan utils.Message, 100),
			Outbox: coordinatorOutput,
		},
	}

	// create shell for ArchetypeMain (that inits Coordinator)
	mainId := uuid.New()
	coordinator.directory[mainId] = utils.Messenger{
		Me:     mainId,
		Inbox:  make(chan utils.Message, 100),
		Outbox: coordinator.messenger.Inbox,
	}

	log.Printf("coordinator id: %s", coordinator.messenger.Me)
	log.Printf("main id: %s", mainId)

	// add the coordinators coordinatorConfigurator to the directory
	coordinator.directory[coordinator.messenger.Me] = coordinator.messenger

	go coordinator.configForwarder()
	return coordinator, coordinator.GetMessenger(mainId)
}

func (c *Coordinator) configForwarder() {
	for config := range c.messenger.Inbox {
		log.Printf("forwarding message %v", config)
		c.directory[config.GetTo()].Inbox <- config
	}
}
