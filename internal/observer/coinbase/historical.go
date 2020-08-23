package coinbase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/johnmillner/robo-macd/internal/observer"
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

func NewTickerFromCandle(raw []float64, productId string) observer.Ticker {
	return observer.Ticker{
		ProductId: productId,
		Price:     raw[4],
		Time:      time.Unix(int64(raw[0]), 0).UTC(),
	}
}

func CreateCandleQuery(observer *Observer) (*url.URL, error) {
	historicalUrl, err := url.Parse(fmt.Sprintf(observer.coinbase.Price.HistoricalPrice.Https, observer.Product))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not populate historical due to %s", err))
	}

	queries := historicalUrl.Query()
	queries.Set("granularity", fmt.Sprintf("%d", int(observer.Granularity.Seconds())))
	queries.Set("start", time.Now().
		Add(-1*observer.Granularity*time.Duration(observer.prices.Capacity+1)).
		In(time.UTC).
		Format(TimeFormat))
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
func (o *Observer) gatherFrameOfHistorical() error {
	log.Printf("gathering candle data for granularity %d seconds", int(o.Granularity.Seconds()))

	historicalUrl, err := CreateCandleQuery(o)
	if err != nil {
		return err
	}

	rawCandles, err := gatherRawCandles(historicalUrl.String())
	if err != nil {
		return err
	}

	o.restartStack()
	for _, raw := range rawCandles {
		o.UpdatePrice(NewTickerFromCandle(raw, o.Product))
	}

	return nil
}

func (o *Observer) PopulateHistorical() {
	for {
		err := o.gatherFrameOfHistorical()
		if err != nil {
			log.Printf("cannnot load hisotical data due to: %s", err)
			continue
		}
		log.Printf("finished historical - sleeping till next update")
		time.Sleep(o.Granularity)
	}

}

func reverseArray(a [][]float64) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
