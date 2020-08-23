package main

import (
	"github.com/johnmillner/robo-macd/internal/observer"
	"github.com/johnmillner/robo-macd/internal/observer/coinbase"
	"github.com/johnmillner/robo-macd/internal/yaml"
	"log"
	"time"
)

type RoboConfig struct {
	MacdCalculator struct {
		Period float64 `yaml:"period"`
		Trend  struct {
			TrendEmaPeriod int `yaml:"trend-ema-period"`
		} `yaml:"trend"`
	} `yaml:"macd-calculator"`
}

func main() {
	coinbaseConfig := coinbase.Coinbase{}
	err := yaml.ParseYaml("configs/coinbase.yaml", &coinbaseConfig)
	if err != nil {
		log.Fatal(err)
	}

	roboConfig := RoboConfig{}
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

	for tickers := range channel {
		log.Printf("%v", tickers)
	}
}
