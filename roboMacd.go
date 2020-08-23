package main

import (
	"github.com/johnmillner/robo-macd/internal/observer"
	"github.com/johnmillner/robo-macd/internal/observer/coinbase"
	"github.com/johnmillner/robo-macd/internal/yaml"
	"log"
	"time"
)

func main() {
	coinbaseConfig := coinbase.Coinbase{}
	err := yaml.ParseYaml("configs/coinbase.yaml", &coinbaseConfig)
	if err != nil {
		log.Fatal(err)
	}

	channel := make(chan []observer.Ticker, 1000)
	coinbase.NewMonitor("BTC-USD", 10*time.Second, 2, &channel, coinbaseConfig).Initialize()

	for tickers := range channel {
		log.Printf("%v", tickers)
	}
}
