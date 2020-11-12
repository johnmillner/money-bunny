package alpaca_wrapper

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"time"
)

type MockedAlpaca struct {
	Bars     func(period time.Duration, symbols []string, limit int) (map[string][]alpaca.Bar, error)
	Calendar func(start, end string) ([]alpaca.CalendarDay, error)
}

func (m MockedAlpaca) GetBars(period time.Duration, symbols []string, limit int) (map[string][]alpaca.Bar, error) {
	return m.Bars(period, symbols, limit)
}

func (m MockedAlpaca) GetCalendar(start, end string) ([]alpaca.CalendarDay, error) {
	return m.Calendar(start, end)
}

func MockGetBars(duration time.Duration, _ []string, _ int) (map[string][]alpaca.Bar, error) {
	bars := map[string][]alpaca.Bar{
		"TSLA": {{
			Time: time.Now().Add(-3 * duration).Round(time.Minute).Unix(),
			Open: 1,
		}, {
			Time: time.Now().Add(-1 * duration).Round(time.Minute).Unix(),
			Open: 3,
		}},
	}

	return bars, nil
}

func MockCalendar(_, _ string) ([]alpaca.CalendarDay, error) {
	return []alpaca.CalendarDay{{
		Date:  time.Now().Format("2006-01-02"),
		Open:  time.Now().Add(-5 * time.Minute).Format("15:04"),
		Close: time.Now().Add(time.Minute).Format("15:04"),
	}}, nil
}
