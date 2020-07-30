package monitor

import (
	"log"
	"time"
)

type priceMonitor struct {
	Product     string
	Granularity time.Duration
	prices      RasterStack
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

// UpdatePrice is responsible for maintaining the priceMonitor's Price Rasterstack
// there are three scenarios that can happen
// - the new ticker is before the next threshold - do nothing
// - the new ticker is between the next threshold and the next next threshold - add
// - the new ticker is after the next next thresh hold - repeat the prior call to fill in the gap
func (monitor *priceMonitor) UpdatePrice(ticker Ticker) {
	peek, err := monitor.prices.Peek(1)

	if err == nil && ticker.Time.Before(peek.Time.Add(monitor.Granularity)) {
		return
	}

	// todo fix this logic for filling in gaps during live price
	if err == nil && ticker.Time.After(peek.Time.Add(monitor.Granularity*2)) {
		midPrice := Ticker{
			ProductId: peek.ProductId,
			Price:     peek.Price,
			Time:      ticker.Time.Add(-1 * monitor.Granularity),
		}
		monitor.UpdatePrice(midPrice)
	}

	ticker.Time = ticker.Time.Round(monitor.Granularity)
	monitor.prices.Push(ticker)
	monitor.channel <- monitor.prices.Raster()
}

func NewMonitor(product string, granularity time.Duration, capacity int, channel *chan []Ticker, coinbase Coinbase) *priceMonitor {
	return &priceMonitor{
		Product:     product,
		Granularity: granularity,
		channel:     *channel,
		coinbase:    coinbase,
		prices:      NewRasterStack(capacity),
	}
}

func (monitor *priceMonitor) Initialize() {
	err := monitor.PopulateHistorical()
	if err != nil {
		log.Printf("%v", err)
	}
	go monitor.PopulateLive()
}
