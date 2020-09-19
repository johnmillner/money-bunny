package main

import (
	"fmt"
	"github.com/johnmillner/robo-macd/internal/macd"
	"github.com/johnmillner/robo-macd/internal/observer"
	"github.com/johnmillner/robo-macd/internal/observer/coinbase"
	"github.com/johnmillner/robo-macd/internal/yaml"
	"log"
	"net/http"
	"time"
)

func main() {
	coinbaseConfig := coinbase.Coinbase{}
	err := yaml.ParseYaml("configs/coinbase.yaml", &coinbaseConfig)
	if err != nil {
		log.Fatal(err)
	}

	roboConfig := macd.RoboConfig{}
	err = yaml.ParseYaml("configs/config.yaml", &roboConfig)
	if err != nil {
		log.Fatal(err)
	}

	channel := make(chan []observer.Ticker, 2*roboConfig.MacdCalculator.Trend.TrendEmaPeriod)
	for _, product := range coinbaseConfig.Price.Products {
		coinbase.NewMonitor(
			product,
			time.Duration(roboConfig.MacdCalculator.Period)*time.Second,
			roboConfig.MacdCalculator.Trend.TrendEmaPeriod,
			&channel,
			coinbaseConfig).
			Initialize()
	}

	var ticker []observer.Ticker

	//todo
	go func() {
		for tickers := range channel {
			ticker = tickers
		}
	}()

	http.HandleFunc("/observe", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", ticker)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
