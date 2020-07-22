package monitor

import (
	"encoding/json"
	"errors"
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
		Time:      time.Unix(int64(raw[0]), 0).UTC(),
	}
}

func (monitor *priceMonitor) PopulateHistorical() error {
	log.Printf("gathering historical data for granularity %f seconds", monitor.Granularity.Seconds())
	historicalUrl, err := url.Parse(fmt.Sprintf("%s/products/%s/candles", monitor.coinbase.Price.HistoricalPriceHttps, monitor.Product))
	if err != nil {
		return errors.New(fmt.Sprintf("could not populate historical due to %s", err))
	}

	queries := historicalUrl.Query()
	queries.Set("granularity", fmt.Sprintf("%d", int(monitor.Granularity.Seconds())))
	queries.Set("start", time.Now().Add(-1*monitor.Granularity*time.Duration(monitor.prices.capacity)).In(time.UTC).String())
	queries.Set("end", time.Now().Add(time.Minute).In(time.UTC).String())
	historicalUrl.RawQuery = queries.Encode()

	response, err := http.Get(historicalUrl.String())
	if err != nil {
		return errors.New(fmt.Sprintf("could not populate historical due to %s", err))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	rawCandles := make([][]float64, 0)
	err = json.Unmarshal(body, &rawCandles)
	if err != nil {
		return errors.New(fmt.Sprintf("could not populate historical due to %s", err))
	}

	reverseArray(rawCandles)

	for _, raw := range rawCandles {
		monitor.UpdatePrice(NewTickerFromCandle(raw, monitor.Product))
	}

	log.Printf("finished historical")
	return nil
}

func reverseArray(a [][]float64) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
