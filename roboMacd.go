package main

import (
	"github.com/johnmillner/robo-macd/internal/monitor"
	"log"
)

func main() {
	channel := make(chan []monitor.Ticker, 1000)
	monitor.NewMonitor("BTC-USD", 60, &channel)

	for tickers := range channel {
		log.Printf("%v", tickers)
	}
}
