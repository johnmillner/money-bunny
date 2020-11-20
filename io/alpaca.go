package io

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/johnmillner/robo-macd/stock"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"log"
	"time"
)

type Alpaca struct {
	Client *alpaca.Client
}

func NewAlpaca() Alpaca {
	return Alpaca{
		Client: alpaca.NewClient(&common.APIKey{
			ID:     viper.GetString("alpaca.key"),
			Secret: viper.GetString("alpaca.secret"),
		})}
}

func (a Alpaca) GetHistoricalStocks(symbols []string, updates chan stock.Stock) map[string]*stock.Stock {
	stocks := make(map[string]*stock.Stock)

	limit := viper.GetInt("trend") + 2
	bars, err := a.Client.ListBars(symbols, alpaca.ListBarParams{
		Timeframe: "1Min",
		Limit:     &limit,
	})

	if err != nil {
		log.Panicf("could not gather historical prices due to %s", err)
	}

	for symbol, bar := range bars {
		stocks[symbol] = stock.NewStock(symbol, bar, updates)
	}

	return stocks
}

func (a Alpaca) OrderBracket(symbol string, qty, takeProfit, stopLoss, stopLimit float64) {
	tp := decimal.NewFromFloat(takeProfit)
	sl := decimal.NewFromFloat(stopLoss)
	st := decimal.NewFromFloat(stopLimit)

	_, err := a.Client.PlaceOrder(
		alpaca.PlaceOrderRequest{
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(qty),
			Side:        "buy",
			Type:        "market",
			TimeInForce: "day",
			OrderClass:  "bracket",
			TakeProfit: &alpaca.TakeProfit{
				LimitPrice: &tp,
			},
			StopLoss: &alpaca.StopLoss{
				StopPrice:  &sl,
				LimitPrice: &st,
			},
		})

	if err != nil {
		log.Printf("could not complete order for %s from alpaca_wrapper due to %s", symbol, err)
	}
}

func (a Alpaca) CountTradesAndOrders() int {
	return len(a.ListPositions()) + len(a.ListOpenOrders())
}

func (a Alpaca) ListOpenOrders() []alpaca.Order {
	open := "open"
	roll := true
	orders, err := a.Client.ListOrders(&open, nil, nil, &roll)
	if err != nil {
		log.Panicf("could not list open orders in account from io")
	}

	return orders
}

func (a Alpaca) ListPositions() []alpaca.Position {
	positions, err := a.Client.ListPositions()
	if err != nil {
		log.Panicf("could not list positions in account from io")
	}

	return positions
}

func (a Alpaca) GetPortfolioValue() float64 {
	account, err := a.Client.GetAccount()
	if err != nil {
		log.Panicf("could not complete portfollio gather from alpaca_wrapper due to %s", err)
	}

	portfolio, _ := account.PortfolioValue.Float64()

	return portfolio
}

func (a Alpaca) LiquidateOldPositions() {
	orders := a.ListOpenOrders()
	for _, order := range orders {
		if order.SubmittedAt.Add(30 * time.Minute).Before(time.Now()) {
			err := a.Client.ClosePosition(order.Symbol)

			if err != nil {
				log.Printf("could not liqudate old position for %s due to %s", order.Symbol, err)
				continue
			}

			log.Printf("liqudated %s since it was too old", order.Symbol)
		}
	}

	time.Sleep(time.Minute)
}
