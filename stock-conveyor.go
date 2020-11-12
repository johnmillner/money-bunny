package main

import (
	"github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	coordinatorLib "github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"
)

func main() {
	// setup coordinator and receive main's configurator
	coordinator, _ := coordinatorLib.InitCoordinator(make(chan utils.Message, 100))

	//initialize the gatherer
	gatherer := gatherers.Gatherer{}.StartUp(coordinator.NewMessenger(gatherers.GathererConfig{
		EquityData: make(chan []gatherers.Equity, 100000),
		Alpaca:     alpaca_wrapper.Alpaca{},
		Symbols:    []string{"TSLA"},
		Period:     time.Minute,
		Limit:      200,
	}))

	// sleep for just a moment to let the gatherer initialize before shutting it down next block
	time.Sleep(time.Second)

	// tell the gatherer to stop as an example
	gatherer.ShutDown()

	// read through the output of the gatherer as an example
	for simpleData := range gatherer.GetMessenger().Get().(gatherers.GathererConfig).EquityData {
		for _, equity := range simpleData {
			log.Printf("%v", equity)
		}
		log.Printf("%d", len(simpleData))
	}

}
