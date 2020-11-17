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
	ManagerData     chan transformers.TransformedData
}

func (m Manager) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting Macd manager %s", messenger.Me)
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

	buy := transformedData[len(transformedData)-1].Time.After(time.Now().Add(-3*time.Minute)) &&
		isBelowTrend(transformedData) &&
		isUpTrend(transformedData) &&
		isPositiveMacdCrossOver(transformedData)

	if !buy {
		return
	}

	log.Printf("start %v, end%v", transformedData[0].Time, transformedData[len(transformedData)-1].Time)
	config.ManagerData <- transformedData[len(transformedData)-1]
}

func isPositiveMacdCrossOver(transformedData []transformers.TransformedData) bool {
	timeStart := float64(transformedData[len(transformedData)-2].Time.Unix())
	timeEnd := float64(transformedData[len(transformedData)-1].Time.Unix())

	macdStart := transformedData[len(transformedData)-2].Macd
	macdEnd := transformedData[len(transformedData)-1].Macd
	signalStart := transformedData[len(transformedData)-2].Signal
	signalEnd := transformedData[len(transformedData)-1].Signal

	ok, intersection := findIntersection(
		point{
			x: timeEnd,
			y: macdEnd,
		},
		point{
			x: timeStart,
			y: macdStart,
		},
		point{
			x: timeEnd,
			y: signalEnd,
		},
		point{
			x: timeStart,
			y: signalStart,
		})

	return ok &&
		intersection.x >= timeStart && // ensure cross over happened in the last sample
		intersection.x <= timeEnd && // ^
		macdEnd > macdStart && // ensure it is a positive cross over event
		intersection.y < 0 // ensure that the crossover happened in negative space
}

type point struct {
	x, y float64
}

func findIntersection(a, b, c, d point) (bool, point) {
	a1 := b.y - a.y
	b1 := a.x - b.x
	c1 := a1*(a.x) + b1*(a.y)

	a2 := d.y - c.y
	b2 := c.x - d.x
	c2 := a2*(c.x) + b2*(c.y)

	determinant := a1*b2 - a2*b1

	if determinant == 0 {
		return false, point{}
	}

	return true, point{
		x: (b2*c1 - b1*c2) / determinant,
		y: (a1*c2 - a2*c1) / determinant,
	}
}

func isBelowTrend(transformedData []transformers.TransformedData) bool {
	return transformedData[len(transformedData)-1].Price < transformedData[len(transformedData)-1].Trend
}

func isUpTrend(transformedData []transformers.TransformedData) bool {
	return transformedData[len(transformedData)-1].Velocity > 0 || transformedData[len(transformedData)-1].Acceleration > 0
}

func (m Manager) ShutDown() {
	m.active = false
	log.Printf("shutting down Macd manager %s", m.Messenger.Me)
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
