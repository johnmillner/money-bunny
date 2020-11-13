package main

import (
	"github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	coordinatorLib "github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"
)

func main() {
	// setup coordinator and receive main's configurator
	coordinator, _ := coordinatorLib.InitCoordinator(make(chan utils.Message, 100))

	equityData := make(chan []gatherers.Equity, 100000)
	macdData := make(chan []transformers.TransformedData, 100000)
	//initialize the gatherer
	_ = gatherers.Gatherer{}.StartUp(coordinator.NewMessenger(gatherers.GathererConfig{
		EquityData: equityData,
		Alpaca:     alpaca_wrapper.Alpaca{},
		Symbols:    []string{"TSLA"},
		Period:     time.Minute,
		Limit:      200,
	}))

	_ = transformers.Transformer{}.StartUp(coordinator.NewMessenger(transformers.Config{
		EquityData:      equityData,
		TransformedData: macdData,
		Fast:            12,
		Slow:            26,
		Signal:          9,
		InTime:          14,
	}))

	// read through the output of the gatherer as an example
	for simpleData := range macdData {
		for _, equity := range simpleData {
			log.Printf("%+v", equity)
		}
		log.Printf("%d", len(simpleData))
	}

}
