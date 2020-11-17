package internal

import (
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"math"
)

type Executor struct {
	Messenger utils.Messenger
	active    bool
}

type Config struct {
	To   uuid.UUID
	From uuid.UUID

	TransformedData chan transformers.TransformedData
	Alpaca          alpaca_wrapper.AlpacaInterface

	Risk, RiskReward float64
}

func (m Executor) StartUp(messenger utils.Messenger) utils.Module {
	log.Printf("starting executor %s", messenger.Me)
	m.Messenger = messenger
	m.active = true

	go func() {
		for m.active {
			if config, ok := (messenger.Get()).(Config); ok {
				for transformedData := range config.TransformedData {
					go m.buy(transformedData, config)
				}
			} else {
				log.Printf("config received by gatherer not understood %v", config)
			}
		}
	}()

	return m
}

func (m Executor) buy(data transformers.TransformedData, config Config) {
	account, err := config.Alpaca.GetAccount()
	if err != nil {
		log.Panicf("could not complete portfollio gather from alpaca_wrapper due to %s", err)
	}
	portfolio, _ := account.PortfolioValue.Float64()
	risk := config.Risk

	takeProfit := data.Price + config.RiskReward*2*data.Atr
	stopLoss := data.Price - 2*data.Atr
	stopLimit := data.Price - 2.5*data.Atr

	log.Printf("%f %v %f %f %f", data.Price, data.Time, takeProfit, stopLoss, stopLimit)

	qty := int(math.Round(portfolio * risk / data.Price))

	order, err := config.Alpaca.BuyBracket(data.Name, qty, takeProfit, stopLoss, stopLimit)
	if err != nil {
		log.Printf("could not complete order for %s from alpaca_wrapper due to %s", data.Name, err)
	}

	log.Printf("ordered %s %+v", data.Name, order)
}

func (m Executor) ShutDown() {
	m.active = false
	log.Printf("shutting down executor %s", m.Messenger.Me)
}

func (m Executor) GetMessenger() utils.Messenger {
	return m.Messenger
}

func (c Config) GetTo() uuid.UUID {
	return c.To
}

func (c Config) GetFrom() uuid.UUID {
	return c.From
}
