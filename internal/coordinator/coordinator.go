package coordinator

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
)

type Coordinator struct {
	directory    map[uuid.UUID]Description
	configurator utils.Configurator
}

type Description struct {
	Archetype    string
	Configurator utils.Configurator
}

func (c *Coordinator) GetConfigurator(id uuid.UUID) utils.Configurator {
	return c.directory[id].Configurator
}

func (c *Coordinator) GetConfigurators(archetype string) []utils.Configurator {
	list := make([]utils.Configurator, 0)
	for _, descriptor := range c.directory {
		if descriptor.Archetype == archetype {
			list = append(list, descriptor.Configurator)
		}
	}

	return list
}

func InitCoordinator(coordinatorOutput chan utils.Config) (Coordinator, utils.Configurator) {
	coordinator := Coordinator{
		directory: make(map[uuid.UUID]Description),
		configurator: utils.Configurator{
			Me:        uuid.New(),
			ConfigIn:  make(chan utils.Config, 100),
			ConfigOut: coordinatorOutput,
		},
	}

	// create shell for ArchetypeMain (that inits Coordinator)
	mainId := uuid.New()
	coordinator.directory[mainId] = Description{
		Archetype: ArchetypeMain,
		Configurator: utils.Configurator{
			Me:        mainId,
			ConfigIn:  make(chan utils.Config, 100),
			ConfigOut: coordinator.configurator.ConfigIn,
		},
	}

	log.Printf("coordinator id: %s", coordinator.configurator.Me)
	log.Printf("main id: %s", mainId)

	// add the coordinators coordinatorConfigurator to the directory
	coordinator.directory[coordinator.configurator.Me] = Description{
		Archetype:    ArchetypeCoordinator,
		Configurator: coordinator.configurator,
	}

	go coordinator.configForwarder(coordinator.configurator)
	return coordinator, coordinator.GetConfigurator(mainId)
}
