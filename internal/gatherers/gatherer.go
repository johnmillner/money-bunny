package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"sort"
	"time"
)

type Gatherer struct {
	equityData   map[string][]Equity
	Configurator utils.Configurator
}

type Equity struct {
	Name      string
	Time      time.Time
	Open      float32
	Close     float32
	Low       float32
	High      float32
	Volume    int32
	generated bool
}

type GathererConfig struct {
	To     uuid.UUID
	From   uuid.UUID
	Active bool

	EquityData chan []Equity
	Client     alpaca.Client

	Symbols []string
	Period  time.Duration
	Limit   int
}

func (c GathererConfig) GetTo() uuid.UUID {
	return c.To
}

func (c GathererConfig) GetFrom() uuid.UUID {
	return c.To
}

func (c GathererConfig) IsActive() bool {
	return c.Active
}

func chunkList(list []timeRange, chunkSize int) [][]timeRange {
	var chunks [][]timeRange
	for i := 0; i < len(list); i += chunkSize {
		end := i + chunkSize
		if end > len(list) {
			end = len(list)
		}

		chunks = append(chunks, list[i:end])
	}

	return chunks
}

type timeRange struct {
	equity string
	start  time.Time
	end    time.Time
}

func (g *Gatherer) gather(config GathererConfig) {
	// determine wanted Time range
	start := time.Now().Add(time.Duration(-1*config.Limit) * config.Period)
	end := time.Now()

	var requestListNew []timeRange
	var requestListHistory []timeRange
	var requestListUpdate []timeRange

	// determine needed data
	for _, symbol := range config.Symbols {
		// if symbol not in map at all - add total timeframe to request list
		if _, ok := g.equityData[symbol]; !ok {
			requestListNew = append(requestListNew, timeRange{
				equity: symbol,
				start:  start,
				end:    end,
			})
			continue
		}

		equityHistory := g.equityData[symbol]

		// determine if we need prior data for this symbol
		if equityHistory[0].Time.After(start) {
			requestListHistory = append(requestListHistory, timeRange{
				equity: symbol,
				start:  start,
				end:    equityHistory[0].Time,
			})
		}

		// determine if we need new data for this symbol
		if equityHistory[len(equityHistory)-1].Time.Before(end) {
			requestListUpdate = append(requestListUpdate, timeRange{
				equity: symbol,
				start:  equityHistory[len(equityHistory)-1].Time,
				end:    end,
			})
		}
	}

	// chunk each list
	requestRanges := chunkList(requestListNew, config.Limit)
	requestRanges = append(requestRanges, chunkList(requestListHistory, config.Limit)...)
	requestRanges = append(requestRanges, chunkList(requestListUpdate, config.Limit)...)

	// request missing gaps in data and add to intermediary map
	results := make(map[string][]alpaca.Bar)
	for _, requestRange := range requestRanges {
		for symbol, bars := range Request(requestRange, config) {
			results[symbol] = append(results[symbol], bars...)
		}
	}

	// transform data into canonical model
	for symbol, bars := range results {
		for _, bar := range bars {
			g.equityData[symbol] = append(g.equityData[symbol], Equity{
				Name:      symbol,
				Time:      bar.GetTime(),
				Open:      bar.Open,
				Close:     bar.Close,
				Low:       bar.Low,
				High:      bar.High,
				Volume:    bar.Volume,
				generated: false,
			})
		}
	}

	// sort data and backfill missing portions of data
	for symbol := range g.equityData {
		sort.Slice(g.equityData[symbol], func(i, j int) bool {
			return g.equityData[symbol][i].Time.Before(g.equityData[symbol][j].Time)
		})

		for i, equity := range g.equityData[symbol] {
			//skip validation of the last segment since forward looking
			if i >= len(g.equityData[symbol])-1 {
				continue
			}

			// if the element is not there - insert it into the array
			if g.equityData[symbol][i+1].Time != equity.Time.Add(config.Period) {
				g.equityData[symbol] = append(g.equityData[symbol], Equity{})
				copy(g.equityData[symbol][i+1:], g.equityData[symbol][i:])
				g.equityData[symbol][i+1] = g.equityData[symbol][i]
			}
		}
	}

	// pacakge data
	for _, symbol := range config.Symbols {
		config.EquityData <- g.equityData[symbol][len(g.equityData[symbol])-config.Limit:]
	}
}

func InitGatherer(configurator utils.Configurator) Gatherer {
	log.Printf("starting to gather with %v", configurator.Get())
	g := Gatherer{
		equityData:   make(map[string][]Equity),
		Configurator: configurator,
	}

	go func() {
		for config := g.Configurator.Get().(GathererConfig); config.Active; config = g.Configurator.Get().(GathererConfig) {
			go g.gather(config)
			time.Sleep(config.Period)
		}

		log.Printf("shutting down fetcher %s", configurator.Me)
	}()

	return g
}
