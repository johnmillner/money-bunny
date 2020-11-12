package alpaca_wrapper

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"log"
	"time"
)

type AlpacaInterface interface {
	GetBars(period time.Duration, symbols []string, limit int) (map[string][]alpaca.Bar, error)
	GetCalendar(start, end string) ([]alpaca.CalendarDay, error)
}

type Alpaca struct {
	client *alpaca.Client
}

func (a Alpaca) getClient() *alpaca.Client {
	if a.client == nil {
		a.client = alpaca.NewClient(common.Credentials())
	}
	return a.client
}

func (a Alpaca) GetBars(period time.Duration, symbols []string, limit int) (map[string][]alpaca.Bar, error) {
	return a.getClient().ListBars(symbols, alpaca.ListBarParams{
		Timeframe: durationToTimeframe(period),
		Limit:     &limit,
	})
}

func (a Alpaca) GetCalendar(start, end string) ([]alpaca.CalendarDay, error) {
	return alpaca.GetCalendar(&start, &end)
}

func durationToTimeframe(dur time.Duration) string {
	switch dur {
	case time.Minute:
		return string(alpaca.Min1)
	case time.Minute * 5:
		return string(alpaca.Min5)
	case time.Minute * 15:
		return string(alpaca.Min15)
	case time.Hour:
		return string(alpaca.Hour1)
	case time.Hour * 24:
		return string(alpaca.Day1)
	default:
		log.Panicf("cannot translate duration given to alpaca_wrapper timeframe, given: %f (in seconds) "+
			"- only acceptable durations are 1min, 5min, 15min, 1day", dur.Seconds())
		return ""
	}
}
