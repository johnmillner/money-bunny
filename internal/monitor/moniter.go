package monitor

import (
	"encoding/json"
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
		LivePriceWs     string `yaml:"live-price-ws"`
		HistoricalPrice struct {
			Https      string    `yaml:"https"`
			Granulates []float64 `yaml:"granularity"`
		} `yaml:"historical-price"`
		Products []string `yaml:"Products"`
	}
}

// UpdatePrice is responsible for maintaining the priceMonitor's Prices
func (monitor *priceMonitor) UpdatePrice(ticker Ticker) {
	peek, err := monitor.prices.Peek()

	//standardize ticker time rounding down to the nearest granularity
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

	j, _ := json.MarshalIndent(monitor.prices.Raster(), "", "  ")
	log.Printf("%v", string(j))
	//broadcast change if the stack is ready for processing
	if monitor.prices.IsFull() {
		monitor.channel <- monitor.prices.Raster()
	}
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
	//determine if we should be reading historical times or live feed times (if granularity compatible with candles)
	if arrayContains(monitor.coinbase.Price.HistoricalPrice.Granulates, monitor.Granularity.Seconds()) {
		go monitor.PopulateHistorical()
	} else {
		go monitor.PopulateLive()
	}
}

func (monitor *priceMonitor) restartStack() {
	monitor.prices = NewRasterStack(monitor.prices.capacity)
}

func arrayContains(array []float64, target float64) bool {
	for _, f := range array {
		if f == target {
			return true
		}
	}
	return false
}
