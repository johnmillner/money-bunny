package io

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

func (a Alpaca) GetMarketTime() (bool, time.Time, time.Time) {
	today := time.Now().Format("2006-01-02")
	times, err := a.Client.GetCalendar(&today, &today)

	if err != nil {
		logrus.
			WithError(err).
			Panic("could not gather today's time")
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		logrus.
			WithError(err).
			Panic("could not parse timezone")
	}
	marketOpen, err := time.ParseInLocation("2006-01-02T15:04", fmt.Sprintf("%sT%s", times[0].Date, times[0].Open), loc)
	if err != nil {
		logrus.
			WithError(err).
			Panic("could not parse time")
	}
	marketClose, err := time.ParseInLocation("2006-01-02T15:04", fmt.Sprintf("%sT%s", times[0].Date, times[0].Close), loc)
	if err != nil {
		logrus.
			WithError(err).
			Panic("could not parse time")
	}

	return today == times[0].Date, marketOpen, marketClose
}

func (a Alpaca) OrderBracket(symbol string, qty, price, takeProfit, stopLoss, stopLimit float64) {
	pp := decimal.NewFromFloat(price)
	tp := decimal.NewFromFloat(takeProfit)
	sl := decimal.NewFromFloat(stopLoss)
	st := decimal.NewFromFloat(stopLimit)

	_, err := a.Client.PlaceOrder(
		alpaca.PlaceOrderRequest{
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(qty),
			Side:        alpaca.Buy,
			Type:        alpaca.Limit,
			TimeInForce: alpaca.Day,
			LimitPrice:  &pp,
			OrderClass:  alpaca.Bracket,
			TakeProfit: &alpaca.TakeProfit{
				LimitPrice: &tp,
			},
			StopLoss: &alpaca.StopLoss{
				StopPrice:  &sl,
				LimitPrice: &st,
			},
		})

	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Error("could not complete  order")
	}
}

func (a Alpaca) OrderLimit(symbol string, qty, limit decimal.Decimal) {

	_, err := a.Client.PlaceOrder(alpaca.PlaceOrderRequest{
		AssetKey:    &symbol,
		Qty:         qty,
		Side:        alpaca.Sell,
		Type:        alpaca.Limit,
		TimeInForce: alpaca.Day,
		LimitPrice:  &limit,
		OrderClass:  alpaca.Simple,
	})

	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Error("could not complete limit sell order")
	}
}

func (a Alpaca) ListOpenOrders() []alpaca.Order {
	open := "open"
	roll := false
	orders, err := a.Client.ListOrders(&open, nil, nil, &roll)
	if err != nil {
		logrus.
			WithError(err).
			Panic("could not list open orders in account")
	}

	return orders
}

func (a Alpaca) GetQuote(symbol string) *alpaca.LastQuoteResponse {
	quote, err := a.Client.GetLastQuote(symbol)

	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Panic("could not get the last quote")
	}

	return quote
}

func (a *Alpaca) ListPositions() []alpaca.Position {
	positions, err := a.Client.ListPositions()

	if err != nil {
		logrus.
			WithError(err).
			Panic("could not list positions")
	}

	return positions
}

func (a *Alpaca) GetAccount() alpaca.Account {
	account, err := a.Client.GetAccount()

	if err != nil {
		logrus.
			WithError(err).
			Panic("could not get the account")
	}

	return *account
}

func (a *Alpaca) Exit() {
	err := a.Client.CancelAllOrders()
	if err != nil {
		logrus.
			WithError(err).
			Error("could not cancel all orders in alpaca")
	}

	err = a.Client.CloseAllPositions()
	if err != nil {
		logrus.
			WithError(err).
			Error("could not close all positions in alpaca")
	}
}
