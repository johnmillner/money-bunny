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
	case message := <-m.Inbox:
		if message.GetTo() != m.Me {
			log.Printf("received message not meant for me: I am %s - message is %v", m.Me, message)
			break
		}

		m.Current = message
	default:
		// return prior message if nothing waiting in the channel to be picked up
	}

	return m.Current
}

func (m Messenger) GetAndBlock() Message {
	message := <-m.Inbox

	if message.GetTo() != m.Me {
		log.Printf("received message not meant for me: I am %s - message is %v", m.Me, message)
		return m.GetAndBlock()
	}

	log.Printf("got message!")
	m.Current = message

	return m.Current
}

func (m *Messenger) Send(message Message) {
	m.Outbox <- message
}
