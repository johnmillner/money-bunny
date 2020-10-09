package gatherers

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	alpaca2 "github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	"log"
	"sync"
	"time"
)

type MarketTimes struct {
	startRange time.Time
	endRange   time.Time
	lock       sync.RWMutex

	marketTimesMap map[time.Time]timeRange
}

type timeRange struct {
	start time.Time
	end   time.Time
}

type Calendar interface {
	getCalendar(start, end string) ([]alpaca.CalendarDay, error)
}

func NewMarketTimes(startRange, endRange time.Time, alpaca alpaca2.AlpacaInterface) *MarketTimes {
	marketTimes := MarketTimes{
		startRange: startRange,
		endRange:   endRange,
		lock:       sync.RWMutex{},
	}

	marketTimes.marketTimesMap = marketTimes.findMarketTimes(startRange, endRange, alpaca)

	return &marketTimes
}

func (m *MarketTimes) IsMarketOpen(current time.Time) bool {
	dateOfTrade := time.Date(
		current.Year(), current.Month(), current.Day(),
		0, 0, 0, 0, time.Local)

	m.lock.RLock()
	if marketTime, ok := m.marketTimesMap[dateOfTrade]; ok {
		// include 9:30 - exclude 16:00
		m.lock.RUnlock()
		return current.After(marketTime.start.Add(-1)) && current.Before(marketTime.end)
	}

	m.lock.RUnlock()
	return false
}

func (m *MarketTimes) findMarketTimes(startTime, endTime time.Time, alpaca alpaca2.AlpacaInterface) map[time.Time]timeRange {
	marketTimesRaw, err := alpaca.GetCalendar(
		fmt.Sprintf("%d-%d-%d", startTime.Year(), startTime.Month(), startTime.Day()),
		fmt.Sprintf("%d-%d-%d", endTime.Year(), endTime.Month(), endTime.Day()))

	if err != nil {
		log.Panicf("could not get calandar dates for market open due to %s", err)
	}

	marketTimes := make(map[time.Time]timeRange)
	waitGroup := sync.WaitGroup{}
	for _, calendarDay := range marketTimesRaw {
		marketOpen, err := time.ParseInLocation("2006-01-0215:04", calendarDay.Date+calendarDay.Open, time.Local)
		marketClose, err := time.ParseInLocation("2006-01-0215:04", calendarDay.Date+calendarDay.Close, time.Local)

		marketTimeRange := timeRange{
			start: marketOpen,
			end:   marketClose,
		}

		if err != nil {
			log.Panicf("could not parse calendar dates due to %s", err)
		}

		waitGroup.Add(1)
		go func(marketTimeRange timeRange) {
			defer waitGroup.Done()
			m.lock.Lock()
			defer m.lock.Unlock()
			marketTimes[time.Date(
				marketTimeRange.start.Year(),
				marketTimeRange.start.Month(),
				marketTimeRange.start.Day(),
				0, 0, 0, 0, time.Local)] = marketTimeRange
		}(marketTimeRange)
	}
	waitGroup.Wait()

	return marketTimes
}
