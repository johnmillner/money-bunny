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

// UpdatePrice is responsible for maintaining the priceMonitor's Prices
func (monitor *priceMonitor) UpdatePrice(ticker Ticker) {
	peek, err := monitor.prices.Peek(1)

	standardizedTime := ticker.Time.Add(-1 * monitor.Granularity / 2).Add(time.Nanosecond).Round(monitor.Granularity)
	//skip time if ticker occurs before the threshold
	if err == nil && standardizedTime.Before(peek.Time.Add(monitor.Granularity)) {
		return
	}

	//recursively back fill prices if a gap is noticed in data is noticed
	if err == nil && !standardizedTime.Before(peek.Time.Add(2*monitor.Granularity)) {
		monitor.UpdatePrice(Ticker{
			ProductId: peek.ProductId,
			Price:     peek.Price,
			Time:      standardizedTime.Add(-1 * monitor.Granularity),
		})
	}

	// add this price - rounded - to the monitor
	ticker.Time = standardizedTime
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
