package coinbase

import (
	"github.com/gorilla/websocket"
	"github.com/johnmillner/robo-macd/internal/observer"
	"log"
	"reflect"
)

type Channel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

type Definition struct {
	Type     string    `json:"type"`
	Channels []Channel `json:"channels"`
}

func connectToCoinbase(coinbase Coinbase) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(coinbase.Price.LivePriceWs, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func subscribeToProductTicker(connection *websocket.Conn, productId string) error {
	subscribe := Definition{
		Type: "subscribe",
		Channels: []Channel{
			{
				Name:       "ticker",
				ProductIds: []string{productId},
			},
		},
	}

	expectedConfirmation := Definition{
		Type: "subscriptions",
		Channels: []Channel{
			{
				Name:       "ticker",
				ProductIds: []string{productId},
			},
		},
	}

	err := connection.WriteJSON(subscribe)

	subscriptionConfirmation := Definition{}
	for !reflect.DeepEqual(subscriptionConfirmation, expectedConfirmation) {
		err = connection.ReadJSON(&subscriptionConfirmation)
		log.Printf("waiting for confirmation")
	}
	log.Printf("subscription confirmed %v", subscriptionConfirmation)

	return err
}

func readTickerSubscription(connection *websocket.Conn, o *Observer) error {
	for {
		ticker := observer.Ticker{}
		err := connection.ReadJSON(&ticker)
		if err != nil {
			return err
		}

		o.UpdatePrice(ticker)
	}
}

func (o *Observer) PopulateLive() {
	for {
		connection, err := connectToCoinbase(o.coinbase)
		if err != nil {
			log.Fatal(err)
		}

		err = subscribeToProductTicker(connection, o.Product)
		if err != nil {
			log.Printf("there was an issue subscribing to %s live prices: %s", o.Product, err)
		}

		err = readTickerSubscription(connection, o)
		if err != nil {
			log.Printf("there was an issue read tickers - restarting connection %s", err)
		}
	}
}
