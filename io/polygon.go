package io

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/johnmillner/robo-macd/stock"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Action struct {
	Action, Params string
}

type Aggregate struct {
	Ev, Sym                            string
	V, Av, Op, Vw, O, C, H, L, A, S, E float64
}

type Quote struct {
	Status, Symbol string
	Last           struct {
		Asksize, Askexchange, Bidsize, Bidexchange int
		Askprice, Bidprice                         float64
	}
}

func LiveUpdates(stocks map[string]*stock.Stock) {
	inbox := make(chan []byte, 100000)
	outbox := make(chan []byte, 10000)

	c, _, err := websocket.DefaultDialer.Dial("wss://socket.polygon.io/stocks", nil)

	if err != nil {
		log.Panicf("could not connect to polygon %v", err)
	}

	auth, _ := json.Marshal(Action{
		Action: "auth",
		Params: viper.GetString("polygon.key"),
	})
	outbox <- auth

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
				//todo recover
				// panic: ERROR: could not receive message from polygon websocket: close 1006 (abnormal closure): unexpected EOF
			}

			inbox <- b
		}
	}()

	subscribeTicker(stocks, outbox)
	go consumeTickers(stocks, inbox)
}

func subscribeTicker(stocks map[string]*stock.Stock, outbox chan []byte) {
	// stream live aggregate data
	for symbol := range stocks {
		subscribe, _ := json.Marshal(Action{
			Action: "subscribe",
			Params: fmt.Sprintf("AM.%s", symbol),
		})

		outbox <- subscribe
	}
}

func consumeTickers(stocks map[string]*stock.Stock, inbox chan []byte) {
	for b := range inbox {
		go func(b []byte) {
			if strings.Contains(string(b), "\"ev\":\"AM\"") {
				var response []Aggregate
				err := json.Unmarshal(b, &response)

				if err != nil {
					log.Printf("ERROR: could not parse aggregate message, %s", err)
				}

				// for each symbol, update macd, trend, vel, acc, atr just the one value
				for _, a := range response {
					stocks[a.Sym].Update(a.C, a.L, a.H)
				}
			} else {
				log.Printf(string(b))
			}
		}(b)
	}
}

func GetQuote(symbol string) Quote {
	response, err := http.Get(
		fmt.Sprintf(
			"https://api.polygon.io/v1/last_quote/stocks/%s?apiKey=%s",
			symbol,
			viper.GetString("polygon.key")))

	if err != nil {
		log.Panicf("could not gather quote for %s from polygon due to %s", symbol, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Panicf("could not read binary stream quote for %s from polygon due to %s", symbol, err)
	}
	var quote Quote
	err = json.Unmarshal(data, &quote)
	if err != nil {
		log.Panicf("could not unmarshall quote for %s from polygon due to %s", symbol, err)
	}

	return quote
}
