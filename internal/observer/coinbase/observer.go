package coinbase

import (
	"github.com/johnmillner/robo-macd/internal/observer"
	"time"
)

type Observer struct {
	Product     string
	Granularity time.Duration
	prices      observer.Ouroboros
	channel     chan []observer.Ticker
	coinbase    Coinbase
}

type Coinbase struct {
	Price struct {
		LivePriceWs     string `yaml:"live-price-ws"`
		HistoricalPrice struct {
			Https      string    `yaml:"https"`
			Granulates []float64 `yaml:"granularity"`
		} `yaml:"historical-price"`
		Products []string `yaml:"products"`
	}
}

// UpdatePrice is responsible for maintaining the Observer's Prices
func (o *Observer) UpdatePrice(ticker observer.Ticker) {
	o.updatePriceParent(ticker, true)
}

func (o *Observer) updatePriceParent(ticker observer.Ticker, parent bool) {

	peek, err := o.prices.Peek()

	//standardize ticker time rounding down to the nearest granularity
	standardizedTime := ticker.Time.Add(-1 * o.Granularity / 2).Add(time.Nanosecond).Round(o.Granularity)

	//skip time if ticker occurs before the threshold
	if err == nil && standardizedTime.Before(peek.Time.Add(o.Granularity)) {
		return
	}

	//recursively back fill prices if a gap is noticed in data is noticed
	if err == nil && !standardizedTime.Before(peek.Time.Add(2*o.Granularity)) {
		o.updatePriceParent(observer.Ticker{
			ProductId: peek.ProductId,
			Price:     peek.Price,
			Time:      standardizedTime.Add(-1 * o.Granularity),
		}, false)
	}

	// add this ticker
	ticker.Time = standardizedTime
	o.prices = o.prices.Push(ticker)

	//broadcast change if the stack is ready for processing - do not post children
	if o.prices.IsFull() && parent {
		o.channel <- o.prices.Raster()
	}
}

func NewMonitor(
	product string,
	granularity time.Duration,
	capacity int,
	channel *chan []observer.Ticker,
	coinbase Coinbase) *Observer {
	return &Observer{
		Product:     product,
		Granularity: granularity,
		channel:     *channel,
		coinbase:    coinbase,
		prices:      observer.NewOuroboros(capacity),
	}
}

func (o *Observer) Initialize() {
	//determine if we should be reading historical times or live feed times (if granularity compatible with candles)
	if arrayContains(o.coinbase.Price.HistoricalPrice.Granulates, o.Granularity.Seconds()) {
		go o.PopulateHistorical()
	} else {
		go o.PopulateLive()
	}
}

func (o *Observer) restartStack() {
	o.prices = observer.NewOuroboros(o.prices.Capacity)
}

func arrayContains(array []float64, target float64) bool {
	for _, f := range array {
		if f == target {
			return true
		}
	}
	return false
}
