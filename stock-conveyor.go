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

	mainConfigurator.SendConfig(coordinator2.InitConfig{
		To:        coordinatorId,
		From:      mainConfigurator.Me,
		Archetype: coordinator2.ArchetypeGatherer,
		InitialConfig: gatherers.GathererConfig{
			To:         coordinatorId,
			From:       mainConfigurator.Me,
			Active:     true,
			SimpleData: make(chan transformers.SimpleData, 100000),
			Client:     *alpacaClient,
			Symbols:    []string{"TSLA", "AAPL"},
			Limit:      5,
			Period:     time.Minute,
		},
	})

	log.Printf("sent out")
	time.Sleep(5 * time.Second)

	initResponse := mainConfigurator.Get()

	log.Printf("got initResponse %v", initResponse)

	mainConfigurator.SendConfig(gatherers.GathererConfig{
		To:     coordinatorId,
		From:   mainConfigurator.Me,
		Active: false,
	})

	gatherer := coordinator.GetConfigurator(initResponse.(coordinator2.InitResponse).Id)

	log.Printf("got gatherer")

	for simpleData := range gatherer.Get().(gatherers.GathererConfig).SimpleData {
		log.Printf("%v", simpleData)
	}

}
