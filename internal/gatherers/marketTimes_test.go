package gatherers

import (
	"errors"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestMarketTimes_IsMarketOpen_DateNotFound(t *testing.T) {
	times := MarketTimes{
		startRange:     time.Time{},
		endRange:       time.Time{},
		lock:           sync.RWMutex{},
		marketTimesMap: make(map[time.Time]timeRange),
	}

	assert.False(t, times.IsMarketOpen(time.Now()),
		"market should be counted as closed since date not in scope")
}

func TestMarketTimes_IsMarketOpen_BeforeMarket(t *testing.T) {
	times := MarketTimes{
		startRange:     time.Time{},
		endRange:       time.Time{},
		lock:           sync.RWMutex{},
		marketTimesMap: make(map[time.Time]timeRange),
	}

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Year(), 0, 0, 0, 0, time.Local)] = timeRange{
		start: time.Now().Add(5 * time.Hour),
		end:   time.Now().Add(10 * time.Hour),
	}

	assert.False(t, times.IsMarketOpen(time.Now()),
		"market should be counted as closed since date occurred before today's market")
}

func TestMarketTimes_IsMarketOpen_AfterMarket(t *testing.T) {
	times := MarketTimes{
		startRange:     time.Time{},
		endRange:       time.Time{},
		lock:           sync.RWMutex{},
		marketTimesMap: make(map[time.Time]timeRange),
	}

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Year(), 0, 0, 0, 0, time.Local)] = timeRange{
		start: time.Now().Add(-5 * time.Hour),
		end:   time.Now().Add(-10 * time.Hour),
	}

	assert.False(t, times.IsMarketOpen(time.Now()),
		"market should be counted as closed since date occurred after today's market")
}

func TestMarketTimes_IsMarketOpen_DuringMarket(t *testing.T) {
	times := MarketTimes{
		startRange:     time.Time{},
		endRange:       time.Time{},
		lock:           sync.RWMutex{},
		marketTimesMap: make(map[time.Time]timeRange),
	}

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)] = timeRange{
		start: time.Now().Add(-5 * time.Hour),
		end:   time.Now().Add(10 * time.Hour),
	}

	assert.True(t, times.IsMarketOpen(time.Now()),
		"market should be counted as open since date occurred during today's market")
}

func TestMarketTimes_IsMarketOpen_DuringLastMinuteOfMarket(t *testing.T) {
	times := MarketTimes{
		startRange:     time.Time{},
		endRange:       time.Time{},
		lock:           sync.RWMutex{},
		marketTimesMap: make(map[time.Time]timeRange),
	}

	now := time.Now().Round(time.Minute)
	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)] = timeRange{
		start: now.Add(-5 * time.Hour),
		end:   now,
	}

	assert.True(t, times.IsMarketOpen(now.Add(-1*time.Minute)),
		"market should be counted as open since date occurred during today's market")
}

func TestMarketTimes_IsMarketOpen_DuringFirstMinuteOfMarket(t *testing.T) {
	times := MarketTimes{
		startRange:     time.Time{},
		endRange:       time.Time{},
		lock:           sync.RWMutex{},
		marketTimesMap: make(map[time.Time]timeRange),
	}

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)] = timeRange{
		start: time.Now(),
		end:   time.Now().Add(5 * time.Hour),
	}

	assert.True(t, times.IsMarketOpen(time.Now()),
		"market should be counted as open since date occurred during today's market")
}

func TestMarketTimes_NewMarketTimes_BadCalendarCall(t *testing.T) {
	times := MarketTimes{}
	assert.Panics(t, func() {
		_ = times.findMarketTimes(time.Now(), time.Now(), func(start, end string) ([]alpaca.CalendarDay, error) {
			return nil, errors.New("something bad happening")
		})
	})
}

func TestMarketTimes_NewMarketTimes_UnableToParseDates(t *testing.T) {
	times := MarketTimes{}
	assert.Panics(t, func() {
		_ = times.findMarketTimes(time.Now(), time.Now(), func(start, end string) ([]alpaca.CalendarDay, error) {
			return []alpaca.CalendarDay{{
				Date:  "asdf",
				Open:  "asdf",
				Close: "asdf",
			}}, nil
		})
	})
}

func TestMarketTimes_NewMarketTimes_GathersRangeExpected(t *testing.T) {
	times := NewMarketTimes(time.Now().Add(-5*24*time.Hour), time.Now())

	counter := 0
	for date, timeRange := range times.marketTimesMap {
		assert.False(t, date.Weekday() == time.Saturday || date.Weekday() == time.Sunday)
		assert.True(t, timeRange.start == time.Date(date.Year(), date.Month(), date.Day(), 9, 30, 0, 0, time.Local))
		assert.True(t, timeRange.end == time.Date(date.Year(), date.Month(), date.Day(), 16, 0, 0, 0, time.Local))
		counter++
	}

	assert.LessOrEqual(t, counter, 5)
}
