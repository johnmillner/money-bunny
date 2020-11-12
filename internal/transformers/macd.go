package transformers

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
)

type MacdTransformer struct {
	Messenger utils.Messenger
	active    bool
}

type MacdConfig struct {
	To   uuid.UUID
	From uuid.UUID

	EquityData chan []gatherers.Equity
	MacdData   chan []Macd

	Twelve, TwentySix, Nine int
}

type Macd struct {
	macd, signal float32
	volume       int32
}

func (m MacdTransformer) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting macd transformer %s", messenger.Me)
	m.Messenger = messenger
	m.active = true

	go func() {
		for m.active {
			if config, ok := (messenger.Get()).(MacdConfig); ok {
				for equities := range config.EquityData {
					go m.transform(equities, config)
				}
			} else {
				log.Printf("config received by gatherer not understood %v", config)
			}
		}
	}()

	return m
}

func (m MacdTransformer) transform(equities []gatherers.Equity, config MacdConfig) {
	config.MacdData <- transformMacd(equities, config.Twelve, config.TwentySix, config.Nine)
}

func transformMacd(equities []gatherers.Equity, twelve, twentySix, nine int) []Macd {
	macdLine := transformMacdLine(equities, twelve, twentySix)
	signal := transformEma(equities, nine)
	volume := equities

	log.Printf("%d", len(macdLine))
	log.Printf("%d", len(signal))
	log.Printf("%d", len(volume))

	macd := make([]Macd, len(macdLine))
	for i := range macdLine {
		macd[i] = Macd{
			macd:   macdLine[i],
			signal: signal[i],
			volume: volume[i].Volume,
		}
	}

	return macd
}

func transformMacdLine(equities []gatherers.Equity, twelve, twentySix int) []float32 {
	twelveEma := transformEma(equities, twelve)
	twentySixEma := transformEma(equities, twentySix)

	trimmedTwelve := twelveEma[twentySix-twelve:]
	macd := make([]float32, len(twelveEma))
	for i := range twentySixEma {
		macd[i] = trimmedTwelve[i] - twentySixEma[i]
	}

	return macd
}

func transformEma(equities []gatherers.Equity, period int) []float32 {
	emaMultiplier := 2 / (float32(period) + 1)

	ema := make([]float32, 1)
	ema[0] = averagePrice(equities, period)

	trimmed := equities[period+1:]

	for i, equity := range trimmed {
		ema = append(
			ema,
			(equity.Close-trimmed[i].Close)*emaMultiplier+trimmed[i].Close)
	}

	return ema
}

func averagePrice(equities []gatherers.Equity, period int) float32 {
	trimmed := equities[:period+1]

	sum := float32(0)
	for _, equity := range trimmed {
		sum += equity.Close
	}

	return sum / float32(len(trimmed))
}

func (m MacdTransformer) ShutDown() {
	m.active = false
	log.Printf("shutting down macd transformer %s", m.Messenger.Me)
}

func (c MacdConfig) GetTo() uuid.UUID {
	return c.To
}

func (c MacdConfig) GetFrom() uuid.UUID {
	return c.From
}

func (m MacdTransformer) GetMessenger() utils.Messenger {
	return m.Messenger
}
