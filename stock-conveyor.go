package main

import (
	"github.com/johnmillner/robo-macd/internal"
	"github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	coordinatorLib "github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/managers"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"
)

func main() {
	// setup coordinator and receive main's configurator
	coordinator, _ := coordinatorLib.InitCoordinator(make(chan utils.Message, 100))

	alpaca := alpaca_wrapper.Alpaca{}
	equityData := make(chan []gatherers.Equity, 1000)
	macdData := make(chan []transformers.TransformedData, 1000)
	managerData := make(chan transformers.TransformedData, 1000)

	assetsRaw, err := alpaca.ListAsserts()
	if err != nil {
		log.Panicf("could not complete asset collection from alpaca_wrapper due to %s", err)
	}
	assets := make([]string, 0)
	for _, asset := range assetsRaw {
		if asset.Tradable && (asset.Exchange == "NYSE" || asset.Exchange == "NASDAQ") {
			assets = append(assets, asset.Symbol)
		}
	}

	//initialize the gatherer
	_ = gatherers.Gatherer{}.StartUp(coordinator.NewMessenger(gatherers.GathererConfig{
		EquityData: equityData,
		Alpaca:     alpaca,
		Symbols:    assets,
		Period:     time.Minute,
		Limit:      1000,
	}))

	_ = transformers.Transformer{}.StartUp(coordinator.NewMessenger(transformers.Config{
		EquityData:      equityData,
		TransformedData: macdData,
		Fast:            12,
		Slow:            26,
		Signal:          9,
		Trend:           200,
		Smooth:          14,
		InTime:          14,
	}))

	_ = managers.Manager{}.StartUp(coordinator.NewMessenger(managers.Config{
		TransformedData: macdData,
		ManagerData:     managerData,
	}))

	_ = internal.Executor{}.StartUp(coordinator.NewMessenger(internal.Config{
		TransformedData: managerData,
		Alpaca:          alpaca,
		Risk:            0.01,
		RiskReward:      1.5,
	}))

	for manager := range managerData {
		log.Printf("manager data %v", manager)
	}
}
