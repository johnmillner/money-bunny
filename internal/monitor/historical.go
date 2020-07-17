package monitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Candle struct {
	Time   time.Time
	Low    float64
	High   float64
	Open   float64
	Close  float64
	Volume float64
}

func NewTickerFromCandle(raw []float64, productId string) Ticker {
	return Ticker{
		ProductId: productId,
		Price:     raw[4],
		Time:      time.Unix(int64(raw[0]), 0),
	}
}

func (monitor *priceMonitor) PopulateHistorical() {
	log.Printf("gathering historical data for granularity %d seconds", monitor.Granularity)
	historicalUrl, err := url.Parse(fmt.Sprintf("%s/products/%s/candles", monitor.coinbase.Price.HistoricalPriceHttps, monitor.Product))
	if err != nil {
		log.Fatal(err)
	}
	historicalUrl.Query().Add("granularity", string(monitor.Granularity))
	response, err := http.Get(historicalUrl.String())
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	rawCandles := make([][]float64, 0)
	err = json.Unmarshal(body, &rawCandles)
	if err != nil {
		log.Fatal(err)
	}

	reverseArray(rawCandles)

	for _, raw := range rawCandles {
		monitor.updatePrice(NewTickerFromCandle(raw, monitor.Product))
	}

	log.Printf("finished historical")
}

func reverseArray(a [][]float64) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
