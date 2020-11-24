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

func NewAlpaca() Alpaca {
	return Alpaca{
		Client: alpaca.NewClient(&common.APIKey{
			ID:     viper.GetString("alpaca.key"),
			Secret: viper.GetString("alpaca.secret"),
		})}
}

func (a Alpaca) GetHistoricalStocks(symbols []string) map[string]*stock.Stock {
	stocks := make(map[string]*stock.Stock)

	limit := viper.GetInt("trend") + viper.GetInt("snapshot-lookback-min") + 2
	bars, err := a.Client.ListBars(symbols, alpaca.ListBarParams{
		Timeframe: "1Min",
		Limit:     &limit,
	})

	if err != nil {
		log.Panicf("could not gather historical prices due to %s", err)
	}

	for symbol, bar := range bars {
		stocks[symbol] = stock.NewStock(symbol, bar)
		stocks[symbol].LogSnapshot("logging", 0, 0, 0, 0)
	}

	return stocks
}

func (a Alpaca) GetMarketTime() (time.Time, time.Time) {
	today := time.Now().Format("2006-01-02")
	times, err := a.Client.GetCalendar(&today, &today)

	if err != nil {
		log.Panicf("could not gather todays time due to %s", err)
	}

	marketOpen, err := time.Parse("2006-01-02T15:04:05", fmt.Sprintf("%sT%s", times[0].Date, times[0].Open))
	marketClose, err := time.Parse("2006-01-02T15:04:05", fmt.Sprintf("%sT%s", times[0].Date, times[0].Close))

	// if markets are not open today set the markets to open in the future and to close in the past
	if today != times[0].Date {
		return time.Now().Add(24 * time.Hour), time.Time{}
	}

	return marketOpen, marketClose
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
	roll := false
	orders, err := a.Client.ListOrders(&open, nil, nil, &roll)
	if err != nil {
		log.Panicf("could not list open orders in account due to %s", err)
		// todo recover
	}

	return orders
}

func (a Alpaca) GetPosition(symbol string) *alpaca.Position {
	position, err := a.Client.GetPosition(symbol)
	if err != nil {
		log.Panicf("could not list positions in account due to %s", err)
		// todo recover
	}

	return position
}

func (a Alpaca) GetOrders(symbol string) []alpaca.Order {
	orders := make([]alpaca.Order, 0)
	for _, order := range a.ListOpenOrders() {
		if order.Symbol == symbol {
			orders = append(orders, order)
		}
	}

	return orders
}

func (a Alpaca) ListPositions() []alpaca.Position {
	positions, err := a.Client.ListPositions()
	if err != nil {
		log.Panicf("could not list positions in account due to %s", err)
		// todo recover
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

func (a Alpaca) LiquidatePosition(order alpaca.Order) {
	err := a.Client.CancelOrder(order.ID)

	if err != nil {
		log.Printf("could not cancel old order for %s due to %s", order.Symbol, err)
		return
	}

	err = a.Client.ClosePosition(order.Symbol)

	if err != nil {
		log.Printf("could not liqudate old position for %s due to %s", order.Symbol, err)
		return
	}
}
