package main

import (
	"github.com/google/uuid"
	coordinator2 "github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
)

func init() {
	// todo
	_ = os.Setenv(common.EnvApiKeyID, "")
	_ = os.Setenv(common.EnvApiSecretKey, "")
	alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
}

func main() {
	alpacaClient := alpaca.NewClient(common.Credentials())

	configOut := make(chan utils.Config, 100)

	coordinatorId := uuid.New()
	coordinator, mainConfigurator := coordinator2.InitCoordinator(configOut)

	gathererId := gatherers.InitGatherer(coordinator.NewConfigurator(gatherers.GathererConfig{
		To:         coordinatorId,
		From:       mainConfigurator.Me,
		Active:     true,
		SimpleData: make(chan transformers.SimpleData, 100000),
		Client:     *alpacaClient,
		Symbols:    []string{"TSLA", "AAPL"},
		Limit:      5,
		Period:     time.Minute,
	}))

	time.Sleep(1 * time.Second)

	mainConfigurator.SendConfig(gatherers.GathererConfig{
		To:     gathererId,
		From:   mainConfigurator.Me,
		Active: false,
	})

	for simpleData := range coordinator.GetConfigurator(gathererId).Get().(gatherers.GathererConfig).SimpleData {
		log.Printf("%v", simpleData)
	}

}
