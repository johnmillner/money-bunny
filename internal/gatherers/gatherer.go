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
	startYear, startMonth, startDay := time.Now().Add(-1 * 5 * period).Add(-1 * 4 * 24 * time.Hour).Date()
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
			start: marketOpen.AddDate(marketDate.Year(), int(marketDate.Month()), marketDate.Day()),
			end:   marketClose.AddDate(marketDate.Year(), int(marketDate.Month()), marketDate.Day()),
		}

		dataPoints += marketClose.Sub(marketOpen).Milliseconds() / period.Milliseconds()
		log.Printf("number of datapoints %d - currently on date: %s", dataPoints, marketDate)
		if dataPoints >= int64(limit) {
			return marketOpen, lastDay, openMarkets
		}
	}

	// if this didnt do what we want - just grab everything! that makes sense...right...right........right?
	// todo please make better
	log.Fatalf("couldnt determine start date")
	return time.Time{}, time.Time{}, nil
}

func (g *Gatherer) gather(config GathererConfig) {
	// determine wanted Time range
	//magic number 4 is a multiplier to create the start time
	start, end, marketTimes := findTimeRange(config.Limit, config.Period, config.Client)

	for key, val := range marketTimes {
		log.Printf("%s, %v", key, val)
	}

	log.Printf("start %s, end %s", start, end)

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
			// ignore this data if its time is outside of market hours
			year, month, date := bar.GetTime().Date()
			dateOfTrade := time.Date(year, month, date, 0, 0, 0, 0, time.Local)

			if bar.GetTime().Before(marketTimes[dateOfTrade].start) {
				log.Printf("ignoring this data becuase its before market hours, %s %s %s", bar.GetTime(), marketTimes[dateOfTrade].start, marketTimes[dateOfTrade].end)
				continue
			}

			if bar.GetTime().After(marketTimes[dateOfTrade].end) {
				log.Printf("ignoring this data becuase its after market hours, %s %s %s", bar.GetTime(), marketTimes[dateOfTrade].start, marketTimes[dateOfTrade].end)
				continue
			}

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

	log.Printf("%v", g.equityData)

	// sort data
	for symbol := range g.equityData {
		sort.Slice(g.equityData[symbol], func(i, j int) bool {
			return g.equityData[symbol][i].Time.Before(g.equityData[symbol][j].Time)
		})
	}

	// backfill missing portions of data
	for symbol := range g.equityData {
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

	log.Printf("%v", g.equityData)

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
