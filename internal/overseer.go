package internal

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/johnmillner/money-bunny/io"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math"
	"time"
)

type Overseer struct {
	a         *io.Alpaca
	p         *io.Polygon
	Positions []alpaca.Position
	Orders    []alpaca.Order
	Account   alpaca.Account
}

func InitOverseer(a *io.Alpaca, p *io.Polygon, out time.Time) *Overseer {
	o := &Overseer{
		a: a,
		p: p,
	}

	go func() {
		for ; time.Now().Before(out); time.Sleep(30 * time.Second) {
			go func() {
				o.Positions = o.a.ListPositions()
				o.Orders = o.a.ListOpenOrders()
				o.Account = o.a.GetAccount()

				logrus.
					WithField("account", o.Account).
					WithField("positions", o.Positions).
					WithField("orders", o.Orders).
					Debug("current status")

				// check old orders
				for _, order := range o.Orders {
					if order.SubmittedAt.Add(time.Duration(viper.GetInt("liquidate-after-min")) * time.Minute).Before(time.Now()) {
						logrus.
							WithField("stock", order.Symbol).
							WithField("submitted", order.SubmittedAt).
							Info("liquidating since it was too old submitted")
						o.liquidate(order.Symbol)
					}
				}
			}()
		}

		// liquidate if near to close
		if time.Now().After(out) {
			for _, position := range o.Positions {
				logrus.
					WithField("stock", position.Symbol).
					Info("liquidating since it's close to market close %v", out)
				o.liquidate(position.Symbol)
			}
		}
	}()

	return o
}

func (o *Overseer) liquidate(symbol string) {
	for _, order := range o.Orders {
		if order.Symbol == symbol {
			err := o.a.Client.CancelOrder(order.ID)

			if err != nil {
				logrus.
					WithField("stock", symbol).
					WithError(err).
					Panic("could not cancel old order")
			}
		}
	}

	for _, position := range o.Positions {
		if position.Symbol == symbol {
			err := o.a.Client.ClosePosition(symbol)

			if err != nil {
				logrus.
					WithField("stock", symbol).
					WithError(err).
					Error("could not liquidate old position for")
			}
		}
	}
}

func (o *Overseer) Manage(stock *Stock) {
	aggregates := o.p.SubscribeTicker(stock.Symbol)

	for aggregate := range aggregates {
		stock = stock.Update(aggregate)

		//check if already own internal
		holding := false
		var hPosition alpaca.Position
		for _, position := range o.Positions {
			if position.Symbol == stock.Symbol {
				holding = true
				hPosition = position
			}
		}

		if holding && FilterByMacdExit(stock) {
			o.liquidate(stock.Symbol)

			price, _ := hPosition.CurrentPrice.Float64()
			qty, _ := hPosition.Qty.Float64()
			go stock.LogSnapshot("selling", price, qty, 0, 0)

			holding = false
		}

		if !holding && FilterByMacdEntry(stock) {
			price, qty, takeProfit, stopLoss, stopLimit := o.getOrderParameters(stock)

			// check that quantity is above zero, there is sufficient volume for the trade, and the risk/reward is favorable
			if qty < 1 {
				logrus.
					WithField("stock", stock.Symbol).
					Debug("skipping due since insufficient capital")
				continue
			}
			if !FilterByVolume(stock, qty) {
				logrus.
					WithField("stock", stock.Symbol).
					Debug("skipping due since insufficient volume")
				continue
			}
			// if maxLoss is less than budget*risk*percentage
			if ok, minRisk, risk := FilterByRiskGoal(o.calculateBudget(), price, stopLoss, qty); !ok {
				logrus.
					WithField("stock", stock.Symbol).
					Debugf("risk not good enough, wanted minimum risk of %f but only has %f", o.calculateBudget(), minRisk, risk)
				continue
			}

			o.a.OrderBracket(stock.Symbol, qty, takeProfit, stopLoss, stopLimit)
			go stock.LogSnapshot("buying", price, qty, takeProfit, stopLoss)
		}
	}
}

func (o *Overseer) getOrderParameters(s *Stock) (float64, float64, float64, float64, float64) {
	budget := o.calculateBudget()

	quote := o.a.GetQuote(s.Symbol)
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

func (o *Overseer) calculateBudget() float64 {
	if len(o.Positions) >= viper.GetInt("max-positions") {
		return 0
	}

	equity, _ := o.Account.Equity.Float64()
	return equity * viper.GetFloat64("margin-multiplier") / viper.GetFloat64("max-positions")
}
