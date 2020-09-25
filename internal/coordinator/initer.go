package coordinator

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
)

type InitConfig struct {
	To            uuid.UUID
	From          uuid.UUID
	Archetype     string
	InitialConfig utils.Config
}

func (i InitConfig) GetTo() uuid.UUID {
	return i.To
}

func (i InitConfig) GetFrom() uuid.UUID {
	return i.From
}

type InitResponse struct {
	To   uuid.UUID
	From uuid.UUID
	Id   uuid.UUID
}

func (i InitResponse) GetTo() uuid.UUID {
	return i.To
}

func (i InitResponse) GetFrom() uuid.UUID {
	return i.From
}

const ArchetypeGatherer = "gatherer"
const ArchetypeCoordinator = "coordinator"
const ArchetypeMain = "main"

func (c Coordinator) initer(configurator utils.Configurator, config InitConfig) uuid.UUID {

	log.Printf("starting to init %v", config)
	id, _ := uuid.NewRandom()

	c.directory[id] = Description{
		Archetype: config.Archetype,
		Configurator: utils.Configurator{
			Me:        id,
			ConfigIn:  make(chan utils.Config, 100),
			ConfigOut: configurator.ConfigIn,
			Config:    config.InitialConfig,
		}}

	switch config.Archetype {
	case ArchetypeGatherer:
		go gatherers.InitGatherer(c.directory[id].Configurator)
	}

	log.Printf("finished initing %s with id %s", config.Archetype, id)
	return id
}
