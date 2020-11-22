package main

import (
	"github.com/johnmillner/robo-macd/config"
	"github.com/johnmillner/robo-macd/io"
	"github.com/johnmillner/robo-macd/stock"
	"github.com/spf13/viper"
	"log"
	"math"
	"time"
)

func main() {
	log.Print("starting robo-scalper")

	// read in configs
	config.Config()

	symbols := viper.GetStringSlice("stocks")

	updates := make(chan stock.Stock, 10000)

	a := io.NewAlpaca()
	stocks := a.GetHistoricalStocks(symbols, updates)
	io.LiveUpdates(stocks)

	go MonitorCurrentPositions(a, stocks)

	for s := range updates {

		if a.CountTradesAndOrders() < 1 &&
			s.IsBelowTrend() &&
			s.IsUpwardsTrend() &&
			s.IsPositiveMacdCrossOver() {

			price, qty, takeProfit, stopLoss, stopLimit := getOrderParameters(s, a)
			a.OrderBracket(s.Symbol, qty, takeProfit, stopLoss, stopLimit)

			s.LogSnapshot("buying", price, qty, takeProfit, stopLoss)

			// time out to prevent double trading
			time.Sleep(30 * time.Second)
		}
	}
}

func getOrderParameters(s stock.Stock, a io.Alpaca) (float64, float64, float64, float64, float64) {
	quote := io.GetQuote(s.Symbol)
	portfolio := a.GetPortfolioValue()
	portfolioRisk := viper.GetFloat64("risk")

	atr := s.Atr[len(s.Atr)-1]
	exposure := portfolio * portfolioRisk

	price := quote.Last.Askprice - (quote.Last.Askprice-quote.Last.Bidprice)/2

	tradeRisk := 2 * atr
	rewardToRisk := viper.GetFloat64("riskReward")
	stopLossMax := viper.GetFloat64("stopLossMax")

	takeProfit := price + rewardToRisk*tradeRisk
	stopLoss := price - tradeRisk
	stopLimit := price - (1+stopLossMax)*tradeRisk

	qty := math.Round(math.Min(exposure/tradeRisk, portfolio/quote.Last.Askprice))
	//ensure we dont go over
	for qty*price > portfolio {
		qty = qty - 1
	}

	return price, qty, takeProfit, stopLoss, stopLimit
}

func MonitorCurrentPositions(a io.Alpaca, stocks map[string]*stock.Stock) {
	time.Sleep(time.Until(time.Now().Round(time.Minute).Add(time.Minute).Add(2 * time.Second)))
	log.Printf("monitoring current orders")

	for {
		go func() {
			orders := a.ListOpenOrders()

			for _, order := range orders {
				// liquidate old orders
				if order.SubmittedAt.Add(time.Duration(viper.GetInt("liquidate-after-min")) * time.Minute).Before(time.Now()) {
					log.Printf("liqudating %s since it was too old", order.Symbol)
					a.LiquidatePosition(order)

				}

				// check if this order should be sold due to macd crossover
				s := stocks[order.Symbol]
				if !s.IsBelowTrend() &&
					!s.IsUpwardsTrend() &&
					s.IsNegativeMacdCrossUnder() {
					log.Printf("liqudating %s since it's macd has crossed over", order.Symbol)
					a.LiquidatePosition(order)
					position := a.GetPosition(order.Symbol)

					price, _ := position.CurrentPrice.Float64()
					qty, _ := position.Qty.Float64()
					s.LogSnapshot("selling", price, qty, 0, 0)
				}
			}
		}()

		time.Sleep(time.Minute)
	}
}
