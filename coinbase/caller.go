package coinbase

import (
	"github.com/johnmillner/robo-macd/internal/config"
)

// Coinbase contains coinbase specific settings and authentication
type Coinbase struct {
	Auth struct {
		Key    string
		Secret string
	}
}

func getCoinbase() (Coinbase, error) {
	coinbase := Coinbase{}
	err := config.GetConfig("../configs/coinbase.yaml", &coinbase)
	return coinbase, err
}
