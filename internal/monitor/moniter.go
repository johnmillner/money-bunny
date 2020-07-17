package monitor

import (
	"container/ring"
	"github.com/johnmillner/robo-macd/internal/config"
	"log"
	"time"
)

type priceMonitor struct {
	Product     string
	Granularity int
	prices      ring.Ring
	channel     chan []Ticker
	coinbase    Coinbase
}

type Ticker struct {
	ProductId string    `json:"product_id"`
	Price     float64   `json:"price,string"`
	Time      time.Time `json:"time"`
}

type Coinbase struct {
	Price struct {
		LivePriceWs          string   `yaml:"live-price-ws"`
		HistoricalPriceHttps string   `yaml:"historical-price-https"`
		ProductIds           []string `yaml:"products"`
	}
}

func (monitor *priceMonitor) updatePrice(ticker Ticker) {
	log.Printf("updating price with ticker %v", ticker)

	//todo intelligently add based on timestamp and granularity
	//monitor.prices.
	//monitor.channel <- monitor.prices
}

func NewMonitor(product string, granularity int, channel *chan []Ticker) *priceMonitor {
	coinbase := Coinbase{}
	err := config.GetConfig("configs\\coinbase.yaml", &coinbase)
	if err != nil {
		log.Fatal(err)
	}

	m := priceMonitor{
		Product:     product,
		Granularity: granularity,
		channel:     *channel,
		coinbase:    coinbase,
	}

	m.PopulateHistorical()
	go m.PopulateLive()

	return &m
}
