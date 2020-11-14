package managers

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"
)

type Manager struct {
	Messenger utils.Messenger
	active    bool
}

type Config struct {
	To   uuid.UUID
	From uuid.UUID

	TransformedData chan []transformers.TransformedData
	ManagerData     chan []ManagerData
}

type ManagerData struct {
	Time                                                           time.Time
	Price, Macd, Signal, Delta, Trend, Velocity, Acceleration, Atr float64
}

func (m Manager) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting Macd transformer %s", messenger.Me)
	m.Messenger = messenger
	m.active = true

	go func() {
		for m.active {
			if config, ok := (messenger.Get()).(Config); ok {
				for transformedData := range config.TransformedData {
					go m.manage(transformedData, config)
				}
			} else {
				log.Printf("config received by gatherer not understood %v", config)
			}
		}
	}()

	return m
}

func (m Manager) manage(transformedData []transformers.TransformedData, config Config) {

}

func isMacdCrossOver(transformedData []transformers.TransformedData, config Config) bool {
	return false
}

func isBelowTrend(transformedData []transformers.TransformedData, config Config) bool {
	return false
}

func isUpTrend(transformedData []transformers.TransformedData, config Config) bool {
	return false
}

func (m Manager) ShutDown() {
	m.active = false
	log.Printf("shutting down Macd transformer %s", m.Messenger.Me)
}

func (m Manager) GetMessenger() utils.Messenger {
	return m.Messenger
}

func (c Config) GetTo() uuid.UUID {
	return c.To
}

func (c Config) GetFrom() uuid.UUID {
	return c.From
}
