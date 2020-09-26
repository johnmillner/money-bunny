package coordinator

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
)

type Coordinator struct {
	directory    map[uuid.UUID]utils.Configurator
	configurator utils.Configurator
}

func (c *Coordinator) GetConfigurator(id uuid.UUID) utils.Configurator {
	return c.directory[id]
}

func (c *Coordinator) NewConfigurator(initialConfig utils.Config) utils.Configurator {
	configurator := utils.Configurator{
		Me:        uuid.New(),
		ConfigIn:  make(chan utils.Config, 100),
		ConfigOut: c.configurator.ConfigIn,
		Config:    initialConfig,
	}

	c.directory[configurator.Me] = configurator
	return configurator
}

func InitCoordinator(coordinatorOutput chan utils.Config) (Coordinator, utils.Configurator) {
	coordinator := Coordinator{
		directory: make(map[uuid.UUID]utils.Configurator),
		configurator: utils.Configurator{
			Me:        uuid.New(),
			ConfigIn:  make(chan utils.Config, 100),
			ConfigOut: coordinatorOutput,
		},
	}

	// create shell for ArchetypeMain (that inits Coordinator)
	mainId := uuid.New()
	coordinator.directory[mainId] = utils.Configurator{
		Me:        mainId,
		ConfigIn:  make(chan utils.Config, 100),
		ConfigOut: coordinator.configurator.ConfigIn,
	}

	log.Printf("coordinator id: %s", coordinator.configurator.Me)
	log.Printf("main id: %s", mainId)

	// add the coordinators coordinatorConfigurator to the directory
	coordinator.directory[coordinator.configurator.Me] = coordinator.configurator

	go coordinator.configForwarder(coordinator.configurator)
	return coordinator, coordinator.GetConfigurator(mainId)
}

func (c *Coordinator) configForwarder(configurator utils.Configurator) {
	for config := range configurator.ConfigIn {
		log.Printf("forwarding config %v+", config)
		c.directory[config.GetTo()].ConfigIn <- config
	}
}
