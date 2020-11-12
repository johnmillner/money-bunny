package coordinator

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoordinator_GetMessenger_ConfigForwarder(t *testing.T) {

	coordinator := Coordinator{
		directory: make(map[uuid.UUID]utils.Messenger),
		messenger: utils.Messenger{
			Me:      uuid.New(),
			Inbox:   make(chan utils.Message, 10),
			Outbox:  make(chan utils.Message, 10),
			Current: nil,
		},
	}

	id := uuid.New()
	messenger := utils.Messenger{
		Me:      id,
		Inbox:   make(chan utils.Message, 10),
		Outbox:  make(chan utils.Message, 10),
		Current: nil,
	}

	coordinator.directory[id] = messenger
	assert.Equal(t, messenger, coordinator.GetMessenger(id))

	go coordinator.configForwarder()

	coordinator.messenger.Inbox <- FakeMessage{to: id}

	assert.Equal(t, FakeMessage{to: id}, <-coordinator.directory[id].Inbox)

}

func TestInitCoordinator(t *testing.T) {
	mainChan := make(chan utils.Message, 10)
	coordinator, main := InitCoordinator(mainChan)

	assert.NotNil(t, coordinator)
	assert.NotNil(t, main)

	assert.Equal(t, main, coordinator.directory[main.Me])
	assert.Equal(t, coordinator.messenger.Inbox, main.Outbox)
	assert.Equal(t, coordinator.messenger, coordinator.directory[coordinator.messenger.Me])

	main.Send(FakeMessage{to: coordinator.messenger.Me})
	assert.Equal(t, FakeMessage{to: coordinator.messenger.Me}, coordinator.messenger.GetAndBlock())

	coordinator.messenger.Inbox <- FakeMessage{to: main.Me}
	assert.Equal(t, FakeMessage{to: main.Me}, main.GetAndBlock())
}

func TestCoordinator_NewMessenger(t *testing.T) {
	mainChan := make(chan utils.Message, 10)
	coordinator, main := InitCoordinator(mainChan)

	messenger := coordinator.NewMessenger(FakeMessage{to: uuid.New()})

	assert.NotNil(t, messenger)

	main.Send(FakeMessage{to: messenger.Me})
	assert.Equal(t, FakeMessage{to: messenger.Me}, messenger.GetAndBlock())

	messenger.Send(FakeMessage{to: main.Me})
	assert.Equal(t, FakeMessage{to: main.Me}, main.GetAndBlock())
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
