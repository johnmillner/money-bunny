package main

import (
	"github.com/google/uuid"
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

	fetcherId, _ := uuid.NewRandom()

	dataOut := make(chan transformers.SimpleData, 100000)
	configManager := utils.ConfigManager{
		Me:        fetcherId,
		ConfigIn:  make(chan utils.ConfigMessage, 100000),
		ConfigOut: make(chan utils.ConfigMessage, 100000),
		Config: gatherers.FetcherConfig{
			To:         fetcherId,
			Active:     true,
			SimpleData: dataOut,
			Client:     *alpacaClient,
			Symbols:    []string{"TSLA", "AAPL"},
			Period:     time.Minute,
		},
	}

	go gatherers.StartFetching(configManager)

	for simpleData := range dataOut {
		log.Printf("%v", simpleData)
	}
}
