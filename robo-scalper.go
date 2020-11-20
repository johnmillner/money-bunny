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

	go a.LiquidateOldPositions()

	for s := range updates {
		if a.CountTradesAndOrders() < 1 &&
			s.IsBelowTrend() &&
			s.IsUpwardsTrend() &&
			s.IsPositiveMacdCrossOver() {

			price, qty, takeProfit, stopLoss, stopLimit := getOrderParameters(s, a)
			a.OrderBracket(s.Symbol, qty, takeProfit, stopLoss, stopLimit)

			logSnapshot(s, price, qty, takeProfit, stopLoss)

			// time out to prevent double trading
			time.Sleep(30 * time.Second)
		}
	}
}

func logSnapshot(s stock.Stock, price, qty, takeProfit, stopLoss float64) {
	p, m, i, t, v, a, r := s.Snapshot()
	log.Printf("ordered %s:\n\t"+
		"price %v\n\t"+
		"macd %v\n\t"+
		"signal %v\n\t"+
		"trend %v\n\t"+
		"vel %v\n\t"+
		"acc %v\n\t"+
		"atr %v\n\t"+
		"maxProfit: %v\n\t"+
		"maxLoss: %v\n\t"+
		"price: %f\n\t"+
		"takeProfit: %f\n\t"+
		"stopLoss: %f\n\t"+
		"qty: %f",
		s.Symbol, p, m, i, t, v, a, r,
		(takeProfit-price)*qty,
		(price-stopLoss)*qty,
		price,
		takeProfit,
		stopLoss,
		qty)
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
