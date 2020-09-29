package gatherers

import (
	"fmt"
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

// will round up to the nearest start of the day
func findTimeRange(limit int, period time.Duration, client alpaca.Client) (time.Time, time.Time, map[time.Time]timeRange) {
	//guess that the max time is less than 5x what we want + plus an additional 4 days
	startYear, startMonth, startDay := time.Now().Add(-1 * 5 * period).Add(-1 * 7 * 24 * time.Hour).Date()
	endYear, endMonth, endDay := time.Now().Date()
	startString := fmt.Sprintf("%d-%d-%d", startYear, startMonth, startDay)
	todayString := fmt.Sprintf("%d-%d-%d", endYear, endMonth, endDay)

	dates, err := client.GetCalendar(&startString, &todayString)

	log.Printf("%v", dates)
	if err != nil {
		log.Fatalf("there was an issue gathering the date %s", err)
	}

	openMarkets := make(map[time.Time]timeRange, 0)
	lastDay := time.Time{}
	dataPoints := int64(0)
	// loop through array in reverse order to see which date fulfills our data requirements
	for i := len(dates) - 1; i >= 0; i-- {
		marketDate, err := time.Parse("2006-01-02", dates[i].Date)
		marketOpen, err := time.Parse("15:04", dates[i].Open)
		marketClose, err := time.Parse("15:04", dates[i].Close)

		// todo dangerous and probably wrong - check with alpaca on calendar timezone to better determine
		marketDate = time.Date(marketDate.Year(), marketDate.Month(), marketDate.Day(), marketDate.Hour(), marketDate.Minute(), marketDate.Second(), marketDate.Nanosecond(), time.Local)
		marketOpen = time.Date(marketDate.Year(), marketDate.Month(), marketDate.Day(), marketOpen.Hour(), marketOpen.Minute(), marketOpen.Second(), marketOpen.Nanosecond(), time.Local)
		marketClose = time.Date(marketDate.Year(), marketDate.Month(), marketDate.Day(), marketClose.Hour(), marketClose.Minute(), marketClose.Second(), marketClose.Nanosecond(), time.Local)

		log.Printf("date: %s", marketDate)
		if err != nil {
			log.Fatalf("could not parse times given from calandar, %s", err)
		}

		if time.Now().After(marketDate) && time.Now().Before(marketOpen) {
			log.Printf("skipping today becuase todays market hasnt opened yet, %s %s", time.Now(), marketDate)
			continue
		}

		// check for last date
		if lastDay.Before(marketClose) {
			lastDay = marketClose
		}

		openMarkets[marketDate] = timeRange{
			start: marketOpen,
			end:   marketClose,
		}

		dataPoints += marketClose.Sub(marketOpen).Milliseconds() / period.Milliseconds()
		log.Printf("number of datapoints %d - currently on date: %s", dataPoints, marketDate)
		if dataPoints >= int64(limit) {
			log.Printf("found range, %s-%s", marketOpen, lastDay)
			for key, value := range openMarkets {
				log.Printf("marketDate avaiable %s %v", key, value)
			}
			return marketOpen, lastDay, openMarkets
		}
	}

	// if this didnt do what we want - just grab everything! that makes sense...right...right........right?
	// todo please make better
	log.Fatalf("couldnt determine start date")
	return time.Time{}, time.Time{}, nil
}

func (g *Gatherer) findRequestRanges(config GathererConfig, start time.Time, end time.Time) [][]timeRange {
	var requestListNew []timeRange
	var requestListHistory []timeRange
	var requestListUpdate []timeRange

	// determine needed data
	for _, symbol := range config.Symbols {
		// if symbol not in map at all - add total timeframe to request list
		if _, ok := g.equityData[symbol]; !ok {
			log.Printf("gathering new data for %s", symbol)
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
			log.Printf("gathering historical data for %s", symbol)
			requestListHistory = append(requestListHistory, timeRange{
				equity: symbol,
				start:  start,
				end:    equityHistory[0].Time,
			})
		}

		// determine if we need new data for this symbol
		if equityHistory[len(equityHistory)-1].Time.Before(end) {
			log.Printf("updating data for %s", symbol)
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

	return requestRanges
}

func (g *Gatherer) gather(config GathererConfig) {
	// determine wanted Time range
	start, end, marketTimes := findTimeRange(config.Limit, config.Period, config.Client)

	log.Printf("start %s, end %s", start, end)

	requestRanges := g.findRequestRanges(config, start, end)

	// request wanted data and add to intermediary map
	results := make(map[string][]alpaca.Bar)
	for _, requestRange := range requestRanges {
		for symbol, bars := range Request(requestRange, config) {
			results[symbol] = append(results[symbol], bars...)
		}
	}

	// transform data into canonical model
	for symbol, bars := range results {
		for _, bar := range bars {
			// ignore this data if its time is outside of market hours
			barTime := bar.GetTime()

			if !isMarketOpen(barTime, marketTimes) {
				log.Printf("ignoring this data becuase its not during market hours, %s", barTime)
				continue
			}

			g.equityData[symbol] = append(g.equityData[symbol], Equity{
				Name:      symbol,
				Time:      barTime,
				Open:      bar.Open,
				Close:     bar.Close,
				Low:       bar.Low,
				High:      bar.High,
				Volume:    bar.Volume,
				generated: false,
			})
		}
	}

	// sort data
	for symbol := range g.equityData {
		sort.Slice(g.equityData[symbol], func(i, j int) bool {
			return g.equityData[symbol][i].Time.Before(g.equityData[symbol][j].Time)
		})
	}

	// back fill missing portions of data
	for symbol := range g.equityData {
		for i := 0; i < len(g.equityData[symbol])-1; i++ {
			// if the expected next time is during market open and
			// the next time is not the expected time - forward fill
			currentTime := g.equityData[symbol][i].Time
			expectedTime := currentTime.Add(config.Period)
			nextTime := g.equityData[symbol][i+1].Time
			if isMarketOpen(expectedTime, marketTimes) && nextTime.After(expectedTime) {

				priorEquity := g.equityData[symbol][i]
				log.Printf("forward fill needed: index: %d, current: %s, expected: %s, received: %s",
					i, g.equityData[symbol][i].Time,
					g.equityData[symbol][i].Time.Add(config.Period),
					g.equityData[symbol][i+1].Time)

				backFill := Equity{
					Name:      priorEquity.Name,
					Time:      expectedTime,
					Open:      priorEquity.Open,
					Close:     priorEquity.Close,
					Low:       priorEquity.Low,
					High:      priorEquity.High,
					Volume:    priorEquity.Volume,
					generated: true,
				}

				g.equityData[symbol] = Insert(g.equityData[symbol], i+1, backFill)
			}
		}
	}

	// package data
	for _, symbol := range config.Symbols {
		log.Printf("%d", len(g.equityData[symbol]))
		config.EquityData <- g.equityData[symbol][len(g.equityData[symbol])-config.Limit:]
	}
}

func isMarketOpen(current time.Time, marketTimes map[time.Time]timeRange) bool {
	dateOfTrade := time.Date(
		current.Year(), current.Month(), current.Day(),
		0, 0, 0, 0, time.Local)

	if _, ok := marketTimes[dateOfTrade]; !ok {
		return false
	}

	// include 9:30 - exclude 16:00
	return current.After(marketTimes[dateOfTrade].start.Add(-1)) && current.Before(marketTimes[dateOfTrade].end)
}

func Insert(slice []Equity, i int, elems ...Equity) []Equity {
	if elems == nil {
		return slice
	}
	if i <= 0 {
		return append(elems, slice...)
	}
	if i >= len(slice) {
		return append(slice, elems...)
	}

	return append(slice[:i], append(elems, slice[i:]...)...)
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
