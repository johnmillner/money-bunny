package internal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
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
		log.Panicf("could not connect to polygon %v", err)
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
				log.Panicf("could not send message %s to polygon %s", string(b), err)
			}
		}
	}()

	// read messages from polygon websocket
	go func() {
		for {
			_, b, err := c.ReadMessage()

			if err != nil {
				log.Panicf("could not receive message from polygon %s", err)
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
					log.Printf("could not unmarshal the status %s due to %s", string(b), err)
				}

				p.Statuses <- s[0]
			} else if strings.Contains(string(b), "\"ev\":\"AM\"") {
				var a Aggregate
				err = json.Unmarshal(b, &a)
				if err != nil {
					log.Printf("could not unmarshal the aggregate %s due to %s", string(b), err)
				}

				p.lock.RLock()
				if channel, ok := p.subscribeMap[a.Sym]; ok {
					channel <- a
				} else {
					log.Printf("could not feather aggregate for %s, becuase no known subscription", a.Sym)
				}
				p.lock.RUnlock()
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

	subscribe, _ := json.Marshal(Action{
		Action: "subscribe",
		Params: fmt.Sprintf("AM.%s", symbol),
	})

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
		log.Panicf("could not gather financials for %s from polygon due to %s", symbol, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Panicf("could not read binary stream financials for %s from polygon due to %s", symbol, err)
	}

	var financials Financials
	err = json.Unmarshal(data, &financials)
	if err != nil {
		log.Panicf("could not unmarshall financials for %s from polygon due to %s", symbol, err)
	}

	if len(financials.Results) < 1 {
		return 0
	}

	return financials.Results[0].MarketCapitalization
}
