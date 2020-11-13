package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"sync"
	"time"
)

type Gatherer struct {
	Messenger utils.Messenger
	active    bool
}

type Equity struct {
	Name                   string
	Time                   time.Time
	Open, Close, Low, High float64
	Volume                 int32
	generated              bool
}

type GathererConfig struct {
	To   uuid.UUID
	From uuid.UUID

	EquityData chan []Equity

	Symbols []string
	Period  time.Duration
	Limit   int

	Alpaca alpaca_wrapper.AlpacaInterface
}

func (g Gatherer) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting gatherer %s", messenger.Me)
	g.Messenger = messenger
	g.active = true

	go func() {
		for g.active {
			if config, ok := (messenger.Get()).(GathererConfig); ok {
				go g.gather(config)
				time.Sleep(config.Period)
			} else {
				log.Printf("config received by gatherer not understood %v", config)
			}
		}
	}()

	return g
}

func (g Gatherer) ShutDown() {
	g.active = false
	close(g.GetMessenger().Get().(GathererConfig).EquityData)
	log.Printf("shutting down gatherer %s", g.Messenger.Me)
}

func (g *Gatherer) gather(config GathererConfig) {
	// chunk symbols into 200 quantities
	for _, symbols := range chunkList(config.Symbols, 200) {
		log.Printf("grabbing chunk of symbols %v", symbols)
		go func(symbols []string) {
			for _, equities := range gatherPage(symbols, config) {
				// send only the requested amount of information
				startingIndex := len(equities) - config.Limit
				if startingIndex < 0 {
					startingIndex = 0
				}

				config.EquityData <- equities[startingIndex:]
			}
		}(symbols)
	}
}

func gatherPage(symbols []string, config GathererConfig) [][]Equity {

	// request the previous 1000 point
	results, err := config.Alpaca.GetBars(config.Period, symbols, 1000)

	if err != nil {
		log.Panicf("could not gather bars from alpaca_wrapper due to %s", err)
	}

	// find when the market is open
	marketTimes := newMarketTimes(
		results[symbols[0]][0].GetTime(),
		results[symbols[0]][len(results[symbols[0]])-1].GetTime(),
		config.Alpaca)

	waitGroup := sync.WaitGroup{}

	// filter out equities that are outside of market hours
	equities := make([][]Equity, 0)
	for symbol, bars := range results {
		waitGroup.Add(1)
		go func(symbol string, bars []alpaca.Bar) {
			defer waitGroup.Done()
			equities = append(equities, filterByMarketOpen(symbol, bars, marketTimes))
		}(symbol, bars)
	}
	waitGroup.Wait()

	// back fill any missing equitiesMap in this range
	for i, equityList := range equities {
		waitGroup.Add(1)
		go func(i int, equityList []Equity) {
			defer waitGroup.Done()
			equities[i] = fillGaps(equityList, marketTimes, config.Period)
		}(i, equityList)
	}
	waitGroup.Wait()

	return equities
}

func fillGaps(equities []Equity, marketTimes *MarketTimes, period time.Duration) []Equity {
	for i := 0; i < len(equities)-1; i++ {
		// if the expected next time is during market open and
		// the next time is not the expected time - forward fill
		currentTime := equities[i].Time
		expectedTime := currentTime.Add(period)
		nextTime := equities[i+1].Time
		if marketTimes.isMarketOpen(expectedTime) && nextTime.After(expectedTime) {
			backFill := Equity{
				Name:      equities[i].Name,
				Time:      expectedTime,
				Open:      equities[i].Open,
				Close:     equities[i].Close,
				Low:       equities[i].Low,
				High:      equities[i].High,
				Volume:    equities[i].Volume,
				generated: true,
			}

			equities = insert(equities, i+1, backFill)
		}
	}

	return equities
}

func filterByMarketOpen(symbol string, bars []alpaca.Bar, marketTimes *MarketTimes) []Equity {
	equities := make([]Equity, 0)
	for _, bar := range bars {
		if marketTimes.isMarketOpen(bar.GetTime()) {
			equities = append(equities, Equity{
				Name:      symbol,
				Time:      bar.GetTime(),
				Open:      float64(bar.Open),
				Close:     float64(bar.Close),
				Low:       float64(bar.Low),
				High:      float64(bar.High),
				Volume:    bar.Volume,
				generated: false,
			})
		}
	}

	return equities
}

func insert(equities []Equity, i int, equity ...Equity) []Equity {
	if equity == nil {
		return equities
	}
	if i <= 0 {
		return append(equity, equities...)
	}
	if i >= len(equities) {
		return append(equities, equity...)
	}

	return append(equities[:i], append(equity, equities[i:]...)...)
}

func chunkList(list []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(list); i += chunkSize {
		end := i + chunkSize
		if end > len(list) {
			end = len(list)
		}

		chunks = append(chunks, list[i:end])
	}

	return chunks
}

func (c GathererConfig) GetTo() uuid.UUID {
	return c.To
}

func (c GathererConfig) GetFrom() uuid.UUID {
	return c.From
}

func (g Gatherer) GetMessenger() utils.Messenger {
	return g.Messenger
}
