package utils

import (
	"github.com/google/uuid"
	"log"
)

type Config interface {
	GetTo() uuid.UUID
	GetFrom() uuid.UUID
}

type Configurator struct {
	Me        uuid.UUID
	ConfigIn  chan Config
	ConfigOut chan Config
	Config    Config
}

func (c Configurator) Get() Config {
	select {
	case config := <-c.ConfigIn:
		if config.GetTo() != c.Me {
			log.Printf("received configuration not meant for me: I am %s - configuration is %v", c.Me, config)
			break
		}

		log.Printf("got update!")
		c.Config = config
	default:
		// return prior config if nothing waiting in the channel to be picked up
		log.Printf("no update")
	}

	return c.Config
}

func (c *Configurator) SendConfig(config Config) {
	c.ConfigOut <- config
}

// todo add endpoint config ability
