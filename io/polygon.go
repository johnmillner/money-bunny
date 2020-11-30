package io

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

type Financials struct {
	Results []struct {
		MarketCapitalization float64
	}
}

func GetMarketCap(symbol string) (float64, error) {
	response, err := http.Get(
		fmt.Sprintf(
			"https://api.polygon.io/v2/reference/financials/%s?limit=1&type=T&apiKey=%s",
			symbol,
			viper.GetString("polygon.key")))

	if err != nil {
		return 0, fmt.Errorf("could not gather financials for %s from polygon due to %s", symbol, err)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("could not read binary stream financials for %s from polygon due to %s", symbol, err)
	}

	var financials Financials
	err = json.Unmarshal(data, &financials)
	if err != nil {
		return 0, fmt.Errorf("could not unmarshall financials for %s from polygon due to %s", symbol, err)
	}

	if len(financials.Results) < 1 {
		return 0, fmt.Errorf("could not gather marketCap for %s since results was empty, raw data: %s", symbol, string(data))
	}

	return financials.Results[0].MarketCapitalization, nil
}
