package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
)

type Action struct {
	Action, Params string
}

type Status struct {
	Ev, Message, To, Status string
}

type Aggregate struct {
	Am, Sym                            string
	A, Av, C, E, Ev, L, B, O, Op, S, Z float64
}

func main() {

	SYMBOLS := []string{ //todo sync to configuration
		"TSLA",
		"AAPL",
	}

	polygonInbox := make(chan []byte, 100000)
	polygonOutbox := make(chan []byte, 10000)

	polygon(polygonOutbox, polygonInbox)

	auth, _ := json.Marshal(Action{
		Action: "auth",
		Params: os.Getenv("POLYGON_KEY"),
	})

	polygonOutbox <- auth

	// gather historical values populate list/map
	// generate initial macd, trend, vel, acc, atr values traditionally

	// stream live aggregate data
	for _, symbol := range SYMBOLS {
		subscribe, _ := json.Marshal(Action{
			Action: "subscribe",
			Params: fmt.Sprintf("AM.%s", symbol),
		})

		polygonOutbox <- subscribe
	}

	// hook up subscription feeds to feed into each symbol
	for b := range polygonInbox {
		log.Printf(string(b))
	}

	// thread each symbol, as each updates, update macd, trend, vel, acc, atr just the one value
	// decide to buy or sell - send through channel

	// sequentially, receive each buy,
	// check if capacity to buy another block ({configured active trades allowed})
	// grab quote, ensure price has not fluctuated more than {configured}
	// decide qty, take-profit, stopLimit, stopLoss and submit

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
