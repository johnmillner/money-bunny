package main

import (
	"github.com/google/uuid"
	coordinator2 "github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
)

func main() {
	alpacaClient := alpaca.NewClient(common.Credentials())

	configOut := make(chan utils.Config, 100)

	coordinatorId := uuid.New()
	coordinator, mainConfigurator := coordinator2.InitCoordinator(configOut)

	gatherer := gatherers.InitGatherer(coordinator.NewConfigurator(gatherers.GathererConfig{
		To:         coordinatorId,
		From:       mainConfigurator.Me,
		Active:     true,
		EquityData: make(chan []gatherers.Equity, 100000),
		Client:     *alpacaClient,
		Symbols:    []string{"TSLA"},
		Limit:      910,
		Period:     time.Minute,
	}))

	time.Sleep(1 * time.Second)

	mainConfigurator.SendConfig(gatherers.GathererConfig{
		To:     gatherer.Configurator.Me,
		From:   mainConfigurator.Me,
		Active: false,
	})

	for simpleData := range gatherer.Configurator.Get().(gatherers.GathererConfig).EquityData {
		for _, equity := range simpleData {
			log.Printf("%v", equity)
		}
		log.Printf("%d", len(simpleData))

	}

}
