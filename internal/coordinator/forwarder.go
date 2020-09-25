package coordinator

import (
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
)

func (c *Coordinator) configForwarder(configurator utils.Configurator) {
	for config := range configurator.ConfigIn {
		switch config.(type) {
		case InitConfig:
			id := c.initer(configurator, config.(InitConfig))
			configurator.SendConfig(InitResponse{
				To:   configurator.Me,
				From: c.configurator.Me,
				Id:   id,
			})
		default:
			log.Printf("forwarding config %v+", config)

			recipientConfigurator := c.directory[config.GetTo()].Configurator
			recipientConfigurator.SendConfig(config)
		}

	}
}
