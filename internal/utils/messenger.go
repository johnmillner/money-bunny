package utils

import (
	"github.com/google/uuid"
	"log"
)

type Message interface {
	GetTo() uuid.UUID
	GetFrom() uuid.UUID
}

type Messenger struct {
	Me      uuid.UUID
	Inbox   chan Message
	Outbox  chan Message
	Current Message
}

type Module interface {
	GetMessenger() Messenger
	StartUp(messenger Messenger) Module
	ShutDown()
}

func (m Messenger) Get() Message {
	select {
	case config := <-m.Inbox:
		if config.GetTo() != m.Me {
			log.Printf("received configuration not meant for me: I am %s - configuration is %v", m.Me, config)
			break
		}

		m.Current = config
	default:
		// return prior config if nothing waiting in the channel to be picked up
	}

	return m.Current
}

func (m Messenger) GetAndBlock() Message {
	config := <-m.Inbox

	if config.GetTo() != m.Me {
		log.Printf("received configuration not meant for me: I am %s - configuration is %v", m.Me, config)
		return m.GetAndBlock()
	}

	log.Printf("got message!")
	m.Current = config

	return m.Current
}

func (m *Messenger) Send(message Message) {
	m.Outbox <- message
}
