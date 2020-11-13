package transformers

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"github.com/markcheno/go-talib"
	"log"
	"time"
)

type Transformer struct {
	Messenger utils.Messenger
	active    bool
}

type Config struct {
	To   uuid.UUID
	From uuid.UUID

	EquityData      chan []gatherers.Equity
	TransformedData chan []TransformedData

	Fast, Slow, Signal, InTime int
}

type TransformedData struct {
	Time                                                         time.Time
	Price, macd, signal, histogram, rsi, aroonUp, aroonDown, atr float64
	volume                                                       int32
}

func (m Transformer) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting macd transformer %s", messenger.Me)
	m.Messenger = messenger
	m.active = true

	go func() {
		for m.active {
			if config, ok := (messenger.Get()).(Config); ok {
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

func (m Transformer) transform(equities []gatherers.Equity, config Config) {
	config.TransformedData <- transformMacd(
		equities,
		config.Fast,
		config.Slow,
		config.Signal,
		config.InTime)
}

func transformMacd(equities []gatherers.Equity, fast, slow, signal, inTime int) []TransformedData {
	closePrices := make([]float64, len(equities))
	LowPrices := make([]float64, len(equities))
	HighPrices := make([]float64, len(equities))
	for i := range equities {
		closePrices[i] = float64(equities[i].Close)
		LowPrices[i] = float64(equities[i].Low)
		HighPrices[i] = float64(equities[i].High)
	}

	macdLine, signalLine, histogram := talib.Macd(closePrices, fast, slow, signal)
	rsi := talib.Rsi(closePrices, inTime)
	aroonUp, aroonDown := talib.Aroon(HighPrices, LowPrices, inTime)
	atr := talib.Atr(HighPrices, LowPrices, closePrices, inTime)

	macd := make([]TransformedData, len(signalLine))
	for i := range signalLine {

		macd[i].Time = equities[i].Time
		macd[i].Price = equities[i].Close
		macd[i].volume = equities[i].Volume
		macd[i].macd = macdLine[i]
		macd[i].signal = signalLine[i]
		macd[i].histogram = histogram[i]
		macd[i].rsi = rsi[i]
		macd[i].aroonUp = aroonUp[i]
		macd[i].aroonDown = aroonDown[i]
		macd[i].atr = atr[i]
	}

	return macd[slow+signal:]
}

func (m Transformer) ShutDown() {
	m.active = false
	log.Printf("shutting down macd transformer %s", m.Messenger.Me)
}

func (c Config) GetTo() uuid.UUID {
	return c.To
}

func (c Config) GetFrom() uuid.UUID {
	return c.From
}

func (m Transformer) GetMessenger() utils.Messenger {
	return m.Messenger
}
