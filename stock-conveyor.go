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
		Symbols:    []string{"TSLA", "AAPL"},
		Limit:      2000,
		Period:     time.Minute,
	}))

	time.Sleep(1 * time.Second)

	mainConfigurator.SendConfig(gatherers.GathererConfig{
		To:     gatherer.Configurator.Me,
		From:   mainConfigurator.Me,
		Active: false,
	})

	for simpleData := range coordinator.GetConfigurator(gatherer.Configurator.Me).Get().(gatherers.GathererConfig).EquityData {
		j := 0
		log.Printf("%d, %v", len(simpleData))
		for i, equity := range simpleData {
			if i == 0 {
				continue
			}

			if equity.Time.Add(-1*time.Minute) != simpleData[i-1].Time {
				log.Printf("sadness - missing data, expected %s got %s", equity.Time.Add(-1*time.Minute), simpleData[i-1].Time)
				j++
			}

		}
		log.Printf("%d", j)
	}

}
