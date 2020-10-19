package utils

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMessenger_Get_defaultsToCurrentValue(t *testing.T) {
	messenger := Messenger{
		Me:      uuid.New(),
		Inbox:   make(chan Message, 100),
		Outbox:  make(chan Message, 100),
		Current: nil,
	}

	assert.Nil(t, messenger.Get())
}

func TestMessenger_Get_wrongAddressIgnore(t *testing.T) {
	messenger := Messenger{
		Me:      uuid.New(),
		Inbox:   make(chan Message, 100),
		Outbox:  make(chan Message, 100),
		Current: nil,
	}

	message := FakeMessage{uuid.New()}

	messenger.Inbox <- message

	assert.Nil(t, messenger.Get())
}

func TestMessenger_Get_retrievesNewValue(t *testing.T) {
	messenger := Messenger{
		Me:      uuid.New(),
		Inbox:   make(chan Message, 100),
		Outbox:  make(chan Message, 100),
		Current: nil,
	}

	message := FakeMessage{messenger.Me}

	messenger.Inbox <- message

	assert.Equal(t, message, messenger.Get())
}

func TestMessenger_GetAndBlock_wrongAddressIgnore(t *testing.T) {
	messenger := Messenger{
		Me:      uuid.New(),
		Inbox:   make(chan Message, 100),
		Outbox:  make(chan Message, 100),
		Current: nil,
	}

	badMessage := FakeMessage{uuid.New()}
	validMessage := FakeMessage{messenger.Me}

	messenger.Inbox <- badMessage
	go func() {
		//valid message
		time.Sleep(time.Second)
		messenger.Inbox <- validMessage
	}()

	start := time.Now()
	assert.Equal(t, validMessage, messenger.GetAndBlock())
	assert.GreaterOrEqual(t, time.Now().Sub(start).Milliseconds(), time.Second.Milliseconds())
}

func TestMessenger_GetAndBlock_retrievesNewValue(t *testing.T) {
	messenger := Messenger{
		Me:      uuid.New(),
		Inbox:   make(chan Message, 100),
		Outbox:  make(chan Message, 100),
		Current: nil,
	}

	message := FakeMessage{messenger.Me}

	go func() {
		time.Sleep(time.Second)
		messenger.Inbox <- message
	}()

	start := time.Now()
	assert.Equal(t, message, messenger.GetAndBlock())
	assert.GreaterOrEqual(t, time.Now().Sub(start).Milliseconds(), time.Second.Milliseconds())
}

func TestMessenger_Send(t *testing.T) {
	messenger := Messenger{
		Me:      uuid.New(),
		Inbox:   make(chan Message, 100),
		Outbox:  make(chan Message, 100),
		Current: nil,
	}

	message := FakeMessage{messenger.Me}

	messenger.Send(message)

	assert.Equal(t, message, <-messenger.Outbox)
}

type FakeMessage struct {
	to uuid.UUID
}

func (f FakeMessage) GetTo() uuid.UUID {
	return f.to
}

func (f FakeMessage) GetFrom() uuid.UUID {
	return uuid.New()
}
