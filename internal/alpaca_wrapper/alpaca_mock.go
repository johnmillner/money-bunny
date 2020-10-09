package alpaca_wrapper

import (
	"errors"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGatherer_DurationToTimeframe(t *testing.T) {
	assert.Equal(t, "1Min", durationToTimeframe(time.Minute))
	assert.Equal(t, "5Min", durationToTimeframe(5*time.Minute))
	assert.Equal(t, "15Min", durationToTimeframe(15*time.Minute))
	assert.Equal(t, "1H", durationToTimeframe(time.Hour))
	assert.Equal(t, "1D", durationToTimeframe(24*time.Hour))

	assert.Panics(t, func() {
		_ = durationToTimeframe(2 * time.Minute)
	})
}

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

func GetCalendarUnparsableDates(start, end string) ([]alpaca.CalendarDay, error) {
	return []alpaca.CalendarDay{{
		Date:  "asdf",
		Open:  "asdf",
		Close: "asdf",
	}}, nil
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

func MockGetBarsFail(_ time.Duration, _ []string, _ int) (map[string][]alpaca.Bar, error) {
	return nil, errors.New("test failure")
}

func MockCalander(_, _ string) ([]alpaca.CalendarDay, error) {
	return []alpaca.CalendarDay{{
		Date:  time.Now().Format("2006-01-02"),
		Open:  time.Now().Add(-5 * time.Minute).Format("15:04"),
		Close: time.Now().Format("15:04"),
	}}, nil
}
