package main

import (
	"encoding/json"
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/gorilla/websocket"
	"github.com/markcheno/go-talib"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type Action struct {
	Action, Params string
}

type Aggregate struct {
	Ev, Sym                            string
	V, Av, Op, Vw, O, C, H, L, A, S, E float64
}

type Stock struct {
	Symbol                             string
	Price, High, Low                   Ouroboros
	Macd, Signal, Trend, Vel, Acc, Atr []float64
	updates                            chan Stock
}

type Quote struct {
	Status, Symbol string
	Last           struct {
		Asksize, Askexchange, Bigsize, Bigexchange int
		Askprice, Bigprice                         float64
		Timestamp                                  time.Time
	}
}

func (s *Stock) update(close, low, high float64) {
	s.Price = s.Price.Push(close)
	s.Low = s.Low.Push(low)
	s.High = s.High.Push(high)

	prices := s.Price.Raster()
	s.Macd, s.Signal, _ = talib.Macd(prices, 12, 26, 9)
	s.Trend, s.Vel, s.Acc = getTrends(prices)
	s.Atr = talib.Atr(s.High.Raster(), s.Low.Raster(), prices, 14)

	s.updates <- *s
}

func main() {

	//determine stocks to choose from

	SYMBOLS := []string{ //todo sync to grab stocks directly meeting criteria
		"TSLA",
		"AAPL",
	}

	alpacaClient := alpaca.NewClient(common.Credentials())

	polygonInbox := make(chan []byte, 100000)
	polygonOutbox := make(chan []byte, 10000)

	polygon(polygonOutbox, polygonInbox)

	polygonKey := os.Getenv("POLYGON_KEY")
	auth, _ := json.Marshal(Action{
		Action: "auth",
		Params: polygonKey,
	})

	polygonOutbox <- auth

	// gather historical values populate list/map
	limit := 360
	bars, err := alpacaClient.ListBars(SYMBOLS, alpaca.ListBarParams{
		Timeframe: "1Min",
		Limit:     &limit,
	})

	if err != nil {
		log.Panicf("could not gather historical prices due to %s", err)
	}

	// generate initial macd, trend, vel, acc, atr values traditionally
	stocks := make(map[string]*Stock)
	updates := make(chan Stock, 10000)
	for symbol, bar := range bars {
		closingPrices := make([]float64, len(bar))
		lowPrices := make([]float64, len(bar))
		highPrices := make([]float64, len(bar))
		for _, b := range bar {
			closingPrices = append(closingPrices, float64(b.Close))
			lowPrices = append(lowPrices, float64(b.Low))
			highPrices = append(highPrices, float64(b.High))
		}

		macd, signal, _ := talib.Macd(closingPrices, 12, 29, 9)

		trend, vel, acc := getTrends(closingPrices)

		atr := talib.Atr(highPrices, lowPrices, closingPrices, 14)

		stocks[symbol] = &Stock{
			Symbol:  symbol,
			Price:   NewOuroboros(closingPrices),
			Low:     NewOuroboros(lowPrices),
			High:    NewOuroboros(highPrices),
			Macd:    macd,
			Signal:  signal,
			Trend:   trend,
			Vel:     vel,
			Acc:     acc,
			Atr:     atr,
			updates: updates,
		}
	}

	// stream live aggregate data
	for _, symbol := range SYMBOLS {
		subscribe, _ := json.Marshal(Action{
			Action: "subscribe",
			Params: fmt.Sprintf("AM.%s", symbol),
		})

		polygonOutbox <- subscribe
	}

	// hook up subscription feeds to feed into each symbol
	go func() {
		for b := range polygonInbox {
			if strings.Contains(string(b), "\"ev\":\"AM\"") {
				var response []Aggregate
				err = json.Unmarshal(b, &response)

				if err != nil {
					log.Printf("ERROR: could not parse aggregate message, %s", err)
				}

				// for each symbol, update macd, trend, vel, acc, atr just the one value
				for _, a := range response {
					stocks[a.Sym].update(a.C, a.L, a.H)
				}
			} else {
				log.Printf(string(b))
			}
		}
	}()

	// decide to buy - send through channel
	for s := range updates {
		positions, err := alpacaClient.ListPositions()
		if err != nil {
			log.Panicf("could not list positions in account from alpaca")
		}

		if len(positions) < 1 && //todo magic number describing number of active trades allowed at one time
			isBelowTrend(s) &&
			isUpTrend(s) &&
			isPositiveMacdCrossOver(s) {

			response, err := http.Get(fmt.Sprintf("https://api.polygon.io/v1/last_quote/stocks/%s?apiKey=%s", s.Symbol, polygonKey))

			if err != nil {
				log.Panicf("could not gather quote for %s from polygon due to %s", s.Symbol, err)
			}

			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Panicf("could not read binary stream quote for %s from polygon due to %s", s.Symbol, err)
			}
			var quote Quote
			err = json.Unmarshal(data, &quote)
			if err != nil {
				log.Panicf("could not unmarshall quote for %s from polygon due to %s", s.Symbol, err)
			}

			account, err := alpacaClient.GetAccount()
			if err != nil {
				log.Panicf("could not complete portfollio gather from alpaca_wrapper due to %s", err)
			}
			portfolio, _ := account.PortfolioValue.Float64()
			risk := 0.01 //todo risk=0.01

			price := quote.Last.Askprice - (quote.Last.Askprice-quote.Last.Bigprice)/2
			atr := s.Atr[len(s.Atr)-1]
			takeProfit := decimal.NewFromFloat(price + (1.5)*2*atr) //todo riskReward=1.5
			stopLoss := decimal.NewFromFloat(price - 2*atr)
			stopLimit := decimal.NewFromFloat(price - 2.5*atr)

			log.Printf("%f %v %v %v %v", price, quote.Last.Timestamp, takeProfit, stopLoss, stopLimit)

			qty := int(math.Round(portfolio * risk / price))

			order, err := alpacaClient.PlaceOrder(
				alpaca.PlaceOrderRequest{
					AssetKey:    &s.Symbol,
					Qty:         decimal.New(int64(qty), 1),
					Side:        "buy",
					Type:        "market",
					TimeInForce: "day",
					OrderClass:  "bracket",
					TakeProfit: &alpaca.TakeProfit{
						LimitPrice: &takeProfit,
					},
					StopLoss: &alpaca.StopLoss{
						LimitPrice: &stopLimit,
						StopPrice:  &stopLoss,
					},
				})

			if err != nil {
				log.Printf("could not complete order for %s from alpaca_wrapper due to %s", s.Symbol, err)
			}

			log.Printf("ordered %s %+v", s.Symbol, order)
		}
	}
}

func polygon(outbox, inbox chan []byte) {
	c, _, err := websocket.DefaultDialer.Dial("wss://socket.polygon.io/stocks", nil)

	if err != nil {
		log.Panicf("could not connect to polygon %v", err)
	}

	// write messages To polygon websocket
	go func() {
		for b := range outbox {
			err = c.WriteMessage(websocket.TextMessage, b)

			if err != nil {
				log.Printf("ERROR: could not send message %s to polygon %v", string(b), err)
			}
		}
	}()

	// read messages from polygon websocket
	go func() {
		for {
			_, b, err := c.ReadMessage()

			if err != nil {
				log.Panicf("ERROR: could not receive message from polygon %+v", err)
				//recover
			}

			inbox <- b
		}
	}()
}

func getTrends(price []float64) ([]float64, []float64, []float64) {
	trend := talib.Ema(price, 200)

	trendVelocity := make([]float64, len(trend))
	for i := range trend {
		if i == 0 || trend[i-1] == 0 {
			continue
		}
		trendVelocity[i] = trend[i] - trend[i-1]
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

	return trend, trendVelocity, trendAcceleration
}

func isPositiveMacdCrossOver(stock Stock) bool {
	macdStart := stock.Macd[len(stock.Macd)-2]
	macdEnd := stock.Macd[len(stock.Macd)-1]
	signalStart := stock.Signal[len(stock.Signal)-2]
	signalEnd := stock.Signal[len(stock.Signal)-1]

	ok, intersection := findIntersection(
		point{
			x: 1,
			y: macdEnd,
		},
		point{
			x: 0,
			y: macdStart,
		},
		point{
			x: 1,
			y: signalEnd,
		},
		point{
			x: 0,
			y: signalStart,
		})

	return ok &&
		intersection.x >= 0 && // ensure cross over happened in the last sample
		intersection.x <= 1 && // ^
		macdEnd > macdStart && // ensure it is a positive cross over event
		intersection.y < 0 // ensure that the crossover happened in negative space
}

type point struct {
	x, y float64
}

func findIntersection(a, b, c, d point) (bool, point) {
	a1 := b.y - a.y
	b1 := a.x - b.x
	c1 := a1*(a.x) + b1*(a.y)

	a2 := d.y - c.y
	b2 := c.x - d.x
	c2 := a2*(c.x) + b2*(c.y)

	determinant := a1*b2 - a2*b1

	if determinant == 0 {
		return false, point{}
	}

	return true, point{
		x: (b2*c1 - b1*c2) / determinant,
		y: (a1*c2 - a2*c1) / determinant,
	}
}

func isBelowTrend(stock Stock) bool {
	return stock.Price.Peek() < stock.Trend[len(stock.Trend)-1]
}

func isUpTrend(stock Stock) bool {
	return stock.Vel[len(stock.Vel)-1] > 0 || stock.Acc[len(stock.Acc)-1] > 0
}
