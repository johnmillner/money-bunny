package main

import (
	"github.com/gorilla/websocket"
	"github.com/johnmillner/robo-macd/internal/config"
	"log"
)

type Channel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

type Subscribe struct {
	Type     string    `json:"type"`
	Channels []Channel `json:"channels"`
}

type Ticker struct {
	ProductId string `json:"product_id"`
	Price     string `json:"price"`
	Time      string `json:"time"`
}

type Coinbase struct {
	Price struct {
		Websocket  string   `yaml:"websocket"`
		ProductIds []string `yaml:"products"`
	}
}

func connectToCoinbase(coinbase Coinbase) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(coinbase.Price.Websocket, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func subscribeToProducts(coinbase Coinbase, connection *websocket.Conn) error {
	subscribe := Subscribe{
		Type: "subscribe",
		Channels: []Channel{
			{
				Name:       "ticker",
				ProductIds: coinbase.Price.ProductIds,
			},
		},
	}

	return connection.WriteJSON(subscribe)
}

func readTickers(connection *websocket.Conn) {

	for {
		ticker := Ticker{}
		err := connection.ReadJSON(&ticker)
		if err != nil {
			log.Printf("failed to send message due to: %s", err)
			break
		}

		log.Printf("%s", ticker)
	}

	log.Printf("closing")
	err := connection.Close()
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	coinbaseConfig := Coinbase{}
	err := config.GetConfig("configs\\coinbase.yaml", &coinbaseConfig)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := connectToCoinbase(coinbaseConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = subscribeToProducts(coinbaseConfig, connection)
	if err != nil {
		log.Fatal(err)
	}

	readTickers(connection)

}
