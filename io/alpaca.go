package io

import (
	"fmt"
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

func NewAlpaca() *Alpaca {
	return &Alpaca{
		Client: alpaca.NewClient(&common.APIKey{
			ID:     viper.GetString("alpaca.key"),
			Secret: viper.GetString("alpaca.secret"),
		})}
}

func (a Alpaca) GetStocks(symbols ...string) []stock.Stock {
	stocks := make([]stock.Stock, 0)

	limit := viper.GetInt("trend") + viper.GetInt("snapshot-lookback-min") + 2
	bars, err := a.Client.ListBars(symbols, alpaca.ListBarParams{
		Timeframe: "1Min",
		Limit:     &limit,
	})

	if err != nil {
		log.Panicf("could not gather historical prices due to %s", err)
	}

	for symbol, bar := range bars {
		if len(bar) < limit {
			continue
		}

		stocks = append(stocks, stock.NewStock(symbol, bar))
	}

	return stocks
}

func (a Alpaca) GetMarketTime() (bool, time.Time, time.Time) {
	today := time.Now().Format("2006-01-02")
	times, err := a.Client.GetCalendar(&today, &today)

	if err != nil {
		log.Panicf("could not gather todays time due to %s", err)
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Panicf("could not parse timezone due to %s", err)
	}
	marketOpen, err := time.ParseInLocation("2006-01-02T15:04", fmt.Sprintf("%sT%s", times[0].Date, times[0].Open), loc)
	if err != nil {
		log.Panicf("could not parse time due to %s", err)
	}
	marketClose, err := time.ParseInLocation("2006-01-02T15:04", fmt.Sprintf("%sT%s", times[0].Date, times[0].Close), loc)
	if err != nil {
		log.Panicf("could not parse time due to %s", err)
	}

	return today == times[0].Date, marketOpen, marketClose
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

func (a Alpaca) ListOpenOrders() []alpaca.Order {
	open := "open"
	roll := false
	orders, err := a.Client.ListOrders(&open, nil, nil, &roll)
	if err != nil {
		log.Panicf("could not list open orders in account due to %s", err)
		// todo recover
	}

	return orders
}

func (a Alpaca) GetBuyingPower() float64 {
	account, err := a.Client.GetAccount()
	if err != nil {
		log.Panicf("could not complete portfollio gather from alpaca_wrapper due to %s", err)
	}

	buyingPower, _ := account.DaytradingBuyingPower.Float64()

	return buyingPower
}

func (a Alpaca) LiquidatePosition(order alpaca.Order) {
	err := a.Client.CancelOrder(order.ID)

	if err != nil {
		log.Panicf("could not cancel old order for %s due to %s", order.Symbol, err)
	}

	err = a.Client.ClosePosition(order.Symbol)

	if err != nil {
		log.Printf("could not liqudate old position for %s due to %s", order.Symbol, err)
	}
}

func (a Alpaca) GetQuote(symbol string) *alpaca.LastQuoteResponse {
	quote, err := a.Client.GetLastQuote(symbol)

	if err != nil {
		log.Panicf("could not get the last quote for %s due to %s", symbol, err)
	}

	return quote
}
