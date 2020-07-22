package main

import (
	"github.com/johnmillner/robo-macd/internal/config"
	"github.com/johnmillner/robo-macd/internal/monitor"
	"log"
	"time"
)

func main() {
	coinbase := monitor.Coinbase{}
	err := config.GetConfig("configs\\coinbase.yaml", &coinbase)
	if err != nil {
		log.Fatal(err)
	}

	channel := make(chan []monitor.Ticker, 1000)
	monitor.NewMonitor("BTC-USD", 10*time.Second, 2, &channel, coinbase).Initialize()

	for tickers := range channel {
		log.Printf("%v", tickers)
	}
}
