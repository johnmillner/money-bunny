package io

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Financials struct {
	Results []struct {
		MarketCapitalization float64
	}
}

type Action struct {
	Action, Params string
}

type Status struct {
	Ev, Status, Message string
}

type Aggregate struct {
	Ev, Sym                      string
	V, Av, Op, Vw, O, C, H, L, A float64
	S, E                         int64
}

type Quote struct {
	Status, Symbol string
	Last           struct {
		Asksize, Askexchange, Bidsize, Bidexchange int
		Askprice, Bidprice                         float64
	}
}

type Polygon struct {
	Statuses      chan Status
	inbox, outbox chan []byte
	c             *websocket.Conn
	subscribeMap  map[string]chan Aggregate
	lock          *sync.RWMutex
}

func InitPolygon() *Polygon {
	c, _, err := websocket.DefaultDialer.Dial("wss://socket.polygon.io/stocks", nil)

	if err != nil {
		logrus.
			WithError(err).
			Panic("could not connect to polygon")
	}

	p := &Polygon{
		inbox:        make(chan []byte, 100000),
		outbox:       make(chan []byte, 100000),
		c:            c,
		Statuses:     make(chan Status, 100000),
		subscribeMap: make(map[string]chan Aggregate),
		lock:         &sync.RWMutex{},
	}

	auth, _ := json.Marshal(Action{
		Action: "auth",
		Params: viper.GetString("polygon.key"),
	})

	p.outbox <- auth

	// write messages To polygon websocket
	go func() {
		for b := range p.outbox {
			err = c.WriteMessage(websocket.TextMessage, b)

			if err != nil {
				logrus.
					WithError(err).
					Panicf("could not send message %s to polygon", string(b))
			}
		}
	}()

	// read messages from polygon websocket
	go func() {
		for {
			_, b, err := c.ReadMessage()

			if err != nil {
				logrus.
					WithError(err).
					Panic("could not receive message from polygon")
			}

			p.inbox <- b
		}
	}()

	// feather inbox to aggregates
	go func() {
		for b := range p.inbox {
			if strings.Contains(string(b), "\"ev\":\"status\"") {
				var s []Status
				err = json.Unmarshal(b, &s)
				if err != nil {
					logrus.
						WithError(err).
						Errorf("could not unmarshal the status %s", string(b))
					continue
				}

				p.Statuses <- s[0]
			} else if strings.Contains(string(b), "\"ev\":\"AM\"") {
				var a []Aggregate
				err = json.Unmarshal(b, &a)
				if err != nil {
					logrus.
						WithError(err).
						Errorf("could not unmarshal the aggregate %s", string(b))
					continue
				}

				for _, aggregate := range a {
					p.lock.RLock()
					if channel, ok := p.subscribeMap[aggregate.Sym]; ok {
						channel <- aggregate
					} else {
						logrus.
							WithError(err).
							WithField("stock", aggregate.Sym).
							Panic("could not feather aggregate because no known subscription")
					}
					p.lock.RUnlock()
				}
			}
		}
	}()

	return p
}

func (p *Polygon) SubscribeTicker(symbol string) chan Aggregate {
	channel := make(chan Aggregate, 1000)

	p.lock.Lock()
	p.subscribeMap[symbol] = channel
	p.lock.Unlock()

	subscribe, err := json.Marshal(Action{
		Action: "subscribe",
		Params: fmt.Sprintf("AM.%s", symbol),
	})

	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Panic("could not marshal subscription")
	}

	p.outbox <- subscribe

	return channel
}

func GetMarketCap(symbol string) float64 {
	response, err := http.Get(
		fmt.Sprintf(
			"https://api.polygon.io/v2/reference/financials/%s?limit=1&type=T&apiKey=%s",
			symbol,
			viper.GetString("polygon.key")))

	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Panic("could not gather financials from polygon")
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Panic("could not read binary stream of financials")
	}

	var financials Financials
	err = json.Unmarshal(data, &financials)
	if err != nil {
		logrus.
			WithError(err).
			WithField("stock", symbol).
			Panic("could not unmarshall financials")
	}

	if len(financials.Results) < 1 {
		return 0
	}

	return financials.Results[0].MarketCapitalization
}
