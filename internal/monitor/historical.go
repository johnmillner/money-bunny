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

var TimeFormat = time.RFC3339

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

func CreateCandleQuery(monitor *priceMonitor) (*url.URL, error) {
	historicalUrl, err := url.Parse(fmt.Sprintf(monitor.coinbase.Price.HistoricalPrice.Https, monitor.Product))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not populate historical due to %s", err))
	}

	queries := historicalUrl.Query()
	queries.Set("granularity", fmt.Sprintf("%d", int(monitor.Granularity.Seconds())))
	queries.Set("start", time.Now().Add(-1*monitor.Granularity*time.Duration(monitor.prices.capacity+1)).In(time.UTC).Format(TimeFormat))
	queries.Set("end", time.Now().Add(time.Minute).In(time.UTC).Format(TimeFormat))
	historicalUrl.RawQuery = queries.Encode()

	return historicalUrl, nil
}

func gatherRawCandles(historicalUrl string) ([][]float64, error) {

	response, err := http.Get(historicalUrl)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unsuccesful call to historical due to %s with url %s", err, historicalUrl))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	rawCandles := make([][]float64, 0)
	err = json.Unmarshal(body, &rawCandles)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not populate historical due to %s with body %s", err, body))
	}

	reverseArray(rawCandles)

	return rawCandles, nil
}
func (monitor *priceMonitor) gatherFrameOfHistorical() error {
	log.Printf("gathering candle data for granularity %d seconds", int(monitor.Granularity.Seconds()))

	historicalUrl, err := CreateCandleQuery(monitor)
	if err != nil {
		return err
	}

	rawCandles, err := gatherRawCandles(historicalUrl.String())
	if err != nil {
		return err
	}

	monitor.restartStack()
	for _, raw := range rawCandles {
		monitor.UpdatePrice(NewTickerFromCandle(raw, monitor.Product))
	}

	return nil
}

func (monitor *priceMonitor) PopulateHistorical() {
	for {
		err := monitor.gatherFrameOfHistorical()
		if err != nil {
			log.Printf("cannnot load hisotical data due to: %s", err)
			continue
		}
		log.Printf("finished historical - sleeping till next update")
		time.Sleep(monitor.Granularity)
	}

}

func reverseArray(a [][]float64) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
