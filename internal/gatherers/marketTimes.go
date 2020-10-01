package gatherers

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
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

func NewMarketTimes(startRange, endRange time.Time) *MarketTimes {
	marketTimes := MarketTimes{
		startRange: startRange,
		endRange:   endRange,
		lock:       sync.RWMutex{},
	}

	marketTimes.marketTimesMap = marketTimes.findMarketTimes(startRange, endRange)

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

func (m *MarketTimes) findMarketTimes(startTime, endTime time.Time) map[time.Time]timeRange {
	startDateString := fmt.Sprintf("%d-%d-%d", startTime.Year(), startTime.Month(), startTime.Day())
	endDateString := fmt.Sprintf("%d-%d-%d", endTime.Year(), endTime.Month(), endTime.Day())

	marketTimesRaw, _ := alpaca.GetCalendar(&startDateString, &endDateString)
	marketTimes := make(map[time.Time]timeRange)
	for _, calendarDay := range marketTimesRaw {
		go func(calendar alpaca.CalendarDay) {
			marketDate, err := time.Parse("2006-01-02", calendar.Date)
			marketOpen, err := time.Parse("15:04", calendar.Open)
			marketClose, err := time.Parse("15:04", calendar.Close)

			if err != nil {
				log.Fatalf("could not parse times given from calandar, %s", err)
			}

			m.lock.Lock()
			marketTimes[time.Date(marketDate.Year(), marketDate.Month(), marketDate.Day(), 0, 0, 0, 0, time.Local)] = timeRange{
				start: time.Date(marketDate.Year(), marketDate.Month(), marketDate.Day(), marketOpen.Hour(), marketOpen.Minute(), marketOpen.Second(), 0, time.Local),
				end:   time.Date(marketDate.Year(), marketDate.Month(), marketDate.Day(), marketClose.Hour(), marketClose.Minute(), marketClose.Second(), 0, time.Local),
			}
			m.lock.Unlock()

		}(calendarDay)
	}

	return marketTimes
}
