package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"
)

type GathererConfig struct {
	To         uuid.UUID
	From       uuid.UUID
	Active     bool
	SimpleData chan transformers.SimpleData
	Client     alpaca.Client
	Symbols    []string
	Period     time.Duration
	Limit      int
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

func gather(config GathererConfig) {

	// make api call
	values, err := config.Client.ListBars(config.Symbols, alpaca.ListBarParams{
		Timeframe: durationToTimeframe(config.Period),
		Limit:     &config.Limit,
	})

	if err != nil {
		log.Printf("could not gather bars from alpaca due to %s", err)
		return
	}

	// translate results into canonical and send out
	for symbol, bars := range values {
		bars := bars
		symbol := symbol
		go func() {
			for _, bar := range bars {
				config.SimpleData <- transformers.SimpleData{
					Product: symbol,
					Time:    bar.GetTime(),
					Price:   bar.Close,
				}
			}
		}()
	}

}

func durationToTimeframe(dur time.Duration) string {
	switch dur {
	case time.Minute:
		return "1Min"
	case time.Minute * 5:
		return "5Min"
	case time.Minute * 15:
		return "15Min"
	case time.Hour * 24:
		return "1D"
	default:
		log.Fatalf("cannot translate duration given to alpaca timeframe, given: %f (in seconds) "+
			"- only acceptable durations are 1min, 5min, 15min, 1day", dur.Seconds())
		return dur.String()
	}
}

func InitGatherer(configurator utils.Configurator) uuid.UUID {
	log.Printf("starting to gather with %v", configurator.Get())
	go func() {
		for config := configurator.Get().(GathererConfig); config.Active; config = configurator.Get().(GathererConfig) {
			go gather(config)
			time.Sleep(config.Period)
		}

		log.Printf("shutting down fetcher %s", configurator.Me)
	}()

	return configurator.Me
}
