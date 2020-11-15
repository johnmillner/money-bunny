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

	Fast, Slow, Signal, Trend, Smooth, InTime int
}

type TransformedData struct {
	Name                                                           string
	Time                                                           time.Time
	Price, Macd, Signal, Delta, Trend, Velocity, Acceleration, Atr float64
}

func (m Transformer) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting Macd transformer %s", messenger.Me)
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
	if equities == nil || len(equities) <= config.Trend+config.Smooth-2 {
		return
	}

	config.TransformedData <- transformMacd(
		equities,
		config.Fast,
		config.Slow,
		config.Signal,
		config.Trend,
		config.Smooth,
		config.InTime)
}

func transformMacd(equities []gatherers.Equity, fast, slow, signal, trend, smooth, inTime int) []TransformedData {
	closePrices := make([]float64, len(equities))
	LowPrices := make([]float64, len(equities))
	HighPrices := make([]float64, len(equities))
	for i := range equities {
		closePrices[i] = equities[i].Close
		LowPrices[i] = equities[i].Low
		HighPrices[i] = equities[i].High
	}

	macdLine, signalLine, histogram := talib.Macd(closePrices, fast, slow, signal)
	trendLine := talib.Ema(closePrices, trend)[trend:]

	trendVelocity := make([]float64, len(trendLine))
	for i := range trendLine {
		if i == 0 || trendLine[i-1] == 0 {
			continue
		}
		trendVelocity[i] = trendLine[i] - trendLine[i-1]
	}
	trendVelocity = trendVelocity[1:]

	trendAcceleration := make([]float64, len(trendVelocity))
	for i := range trendVelocity {
		if i == 0 || trendVelocity[i-1] == 0 {
			continue
		}
		trendAcceleration[i] = trendVelocity[i] - trendVelocity[i-1]
	}
	trendAcceleration = trendAcceleration[1:]

	trendVelocity = talib.Ema(trendVelocity, smooth)
	trendAcceleration = talib.Ema(trendAcceleration, smooth)

	atr := talib.Atr(HighPrices, LowPrices, closePrices, inTime)

	macd := make([]TransformedData, len(signalLine))
	for i := range signalLine {
		macd[i].Name = equities[i].Name
		macd[i].Time = equities[i].Time
		macd[i].Price = equities[i].Close
		macd[i].Macd = macdLine[i]
		macd[i].Signal = signalLine[i]
		macd[i].Delta = histogram[i]
		macd[i].Atr = atr[i]

		if i >= trend {
			macd[i].Trend = trendLine[i-trend]
		}
		if i >= trend+1 {
			macd[i].Velocity = trendVelocity[i-trend-1]
		}
		if i >= trend+2 {
			macd[i].Acceleration = trendAcceleration[i-trend-2]
		}
	}

	return macd[trend+smooth+1:]
}

func (m Transformer) ShutDown() {
	m.active = false
	log.Printf("shutting down Macd transformer %s", m.Messenger.Me)
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
