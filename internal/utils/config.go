package utils

import (
	"github.com/google/uuid"
	"log"
)

type ConfigMessage interface {
	GetTo() uuid.UUID
	IsActive() bool
}

type ConfigManager struct {
	Me        uuid.UUID
	ConfigIn  chan ConfigMessage
	ConfigOut chan ConfigMessage
	Config    ConfigMessage
}

func (c ConfigManager) Check() ConfigMessage {
	select {
	case config := <-c.ConfigIn:
		if config.GetTo() != c.Me {
			log.Printf("received configuration not meant for me: I am %s - configuration is %v", c.Me, config)
			break
		}

		c.Config = config
	default:
		// return prior config if nothing waiting in the channel to be picked up
	}

	return c.Config
}

func (c *ConfigManager) SendConfig(config ConfigMessage) {
	c.ConfigOut <- config
}
