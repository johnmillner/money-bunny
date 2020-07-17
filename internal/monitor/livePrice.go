package monitor

import (
	"github.com/gorilla/websocket"
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

func connectToCoinbase(coinbase Coinbase) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(coinbase.Price.LivePriceWs, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func subscribeToProductTicker(connection *websocket.Conn, productId string) error {
	subscribe := Subscribe{
		Type: "subscribe",
		Channels: []Channel{
			{
				Name:       "ticker",
				ProductIds: []string{productId},
			},
		},
	}

	return connection.WriteJSON(subscribe)
}

func readTickerSubscription(connection *websocket.Conn, m *priceMonitor) {

	for {
		ticker := Ticker{}
		err := connection.ReadJSON(&ticker)
		if err != nil {
			log.Printf("failed to send message due to: %s", err)
			break
		}

		m.updatePrice(ticker)
	}

	log.Printf("closing")
	err := connection.Close()
	if err != nil {
		log.Fatal(err)
	}

}

func (monitor *priceMonitor) PopulateLive() {
	connection, err := connectToCoinbase(monitor.coinbase)
	if err != nil {
		log.Fatal(err)
	}

	err = subscribeToProductTicker(connection, monitor.Product)
	if err != nil {
		log.Fatal(err)
	}

	readTickerSubscription(connection, monitor)
}
