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
	coordinator, mainConfigurator := coordinatorLib.InitCoordinator(make(chan utils.Config, 100))

	//initialize the gatherer
	gatherer := gatherers.InitGatherer(coordinator.NewConfigurator(gatherers.GathererConfig{
		Active:     true,
		EquityData: make(chan []gatherers.Equity, 100000),
		Alpaca:     alpaca_wrapper.Alpaca{},
		Period:     0,
		Limit:      0,
	}))

	// sleep for just a moment to let the gatherer initialize before shutting it down next block
	time.Sleep(time.Nanosecond)

	// tell the gatherer to stop as an example
	mainConfigurator.SendConfig(gatherers.GathererConfig{
		To:     gatherer.Configurator.Me,
		From:   mainConfigurator.Me,
		Active: false,
	})

	// read through the output of the gatherer as an example
	for simpleData := range gatherer.Configurator.Get().(gatherers.GathererConfig).EquityData {
		for _, equity := range simpleData {
			log.Printf("%v", equity)
		}
		log.Printf("%d", len(simpleData))
	}

}
