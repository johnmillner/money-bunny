package main

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/johnmillner/robo-macd/config"
	"github.com/johnmillner/robo-macd/io"
	"github.com/johnmillner/robo-macd/stock"
	"github.com/spf13/viper"
	"log"
	"math"
	"runtime/debug"
	"time"
)

func main() {
	log.Print("starting money bunny")

	// read in configs
	config.Config()

	recovery(time.Now(), func() {
		a := io.NewAlpaca()

		today, opens, closes := a.GetMarketTime()

		if !today {
			log.Printf("market is not open today, exiting for the day")
			return
		}

		in := opens.Add(time.Duration(viper.GetInt("trade-after-open-min")) * time.Minute)
		out := closes.Add(-1 * time.Duration(viper.GetInt("close-before-close-min")) * time.Minute)

		if time.Now().Before(in) {
			log.Printf("market has not opened for today yet, waiting until %s", in.String())
			time.Sleep(in.Sub(time.Now()))
		}

		// wait for the next start of the minute
		if time.Now().Second() != 0 {
			startOfMinute := time.Now().Round(time.Minute)
			if startOfMinute.Before(time.Now()) {
				startOfMinute = startOfMinute.Add(time.Minute)
			}

			log.Printf("sleeping until %s", startOfMinute.String())
			time.Sleep(startOfMinute.Sub(time.Now()))
		}

		for ; out.After(time.Now()); time.Sleep(time.Minute) {
			log.Printf("starting buy sell routine")
			go sell(a, out)
			go buy(a)
		}

		log.Printf("market has closed for today, exiting for the day")
	})
}

func buy(a *io.Alpaca) {
	start := time.Now()

	symbols := io.FilterByTradability(a)
	symbols = io.FilterByRisk(a, symbols)

	stocks := a.GetStocks(symbols...)
	stocks = io.FilterByBuyable(stocks...)

	budget := calculateBudget(a, len(stocks))

	for _, potential := range stocks {
		if potential.IsReadyToBuy() {
			price, qty, takeProfit, stopLoss, stopLimit := getOrderParameters(potential, a, budget)
			// todo potentially further filter by market cap and volume

			if qty < 1 {
				go potential.LogSnapshot("skipping", price, qty, takeProfit, stopLoss)
				continue
			}

			a.OrderBracket(potential.Symbol, qty, takeProfit, stopLoss, stopLimit)
			go potential.LogSnapshot("buying", price, qty, takeProfit, stopLoss)
		}
	}

	if time.Now().Sub(start).Seconds() > viper.GetFloat64("buy-sla-sec") {
		log.Printf("total cycle for buying took longer than sla of %f at %s", viper.GetFloat64("buy-sla-sec"), time.Now().Sub(start).String())
	}
}

func calculateBudget(a *io.Alpaca, eligibleStocks int) float64 {
	return a.GetSpendableAmount() / float64(eligibleStocks)
}

func sell(a *io.Alpaca, out time.Time) {
	start := time.Now()

	openOrders := a.ListOpenOrders()
	symbols := make([]string, 0)
	symbolOrder := make(map[string]alpaca.Order)

	for _, order := range openOrders {
		// sell all orders if close to marketClose
		if out.Before(time.Now()) {
			log.Printf("liqudating %s since it's close to market close %v current time %v",
				order.Symbol,
				out,
				time.Now().UTC())
			a.LiquidatePosition(order)
			continue
		}

		// remove old order/positions
		if order.SubmittedAt.Add(time.Duration(viper.GetInt("liquidate-after-min")) * time.Minute).Before(time.Now()) {
			log.Printf("liqudating %s since it was too old submitted at %v current time %v",
				order.Symbol,
				order.SubmittedAt,
				time.Now().UTC())
			a.LiquidatePosition(order)
			continue
		}

		symbolOrder[order.Symbol] = order
		symbols = append(symbols, order.Symbol)
	}

	for _, update := range a.GetStocks(symbols...) {
		if update.IsReadyToSell() {
			qty, _ := symbolOrder[update.Symbol].Qty.Float64()
			go update.LogSnapshot("selling", update.Price.Peek(), qty, 0, 0)
			a.LiquidatePosition(symbolOrder[update.Symbol])
		}
	}

	if time.Now().Sub(start).Seconds() > viper.GetFloat64("sell-sla-sec") {
		log.Printf("total cycle for selling took longer than sla of %f at %s", viper.GetFloat64("buy-sla-sec"), time.Now().Sub(start).String())
	}
}

func getOrderParameters(s stock.Stock, a *io.Alpaca, budget float64) (float64, float64, float64, float64, float64) {
	quote := a.GetQuote(s.Symbol)
	exposure := budget * viper.GetFloat64("risk")
	price := float64(quote.Last.AskPrice - (quote.Last.AskPrice-quote.Last.BidPrice)/2)

	tradeRisk := viper.GetFloat64("stop-loss-atr-ratio") * s.Atr[len(s.Atr)-1]
	rewardToRisk := viper.GetFloat64("risk-reward")
	stopLossMax := viper.GetFloat64("stop-loss-max")

	takeProfit := price + (rewardToRisk * tradeRisk)
	stopLoss := price - tradeRisk
	stopLimit := price - (1+stopLossMax)*tradeRisk

	qty := math.Round(exposure / tradeRisk)

	//ensure we dont go over
	for qty*price > budget {
		qty = qty - 1
	}

	return price, qty, takeProfit, stopLoss, stopLimit
}

func recovery(start time.Time, f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from panic %v", err)
			debug.PrintStack()
			if start.Add(time.Duration(viper.GetInt("recover-frequency-min")) * time.Minute).After(time.Now()) {
				log.Panicf("too many panics - will not recover due to %v", err)
				return
			}

			go recovery(time.Now(), f)
		}
	}()

	f()
}
