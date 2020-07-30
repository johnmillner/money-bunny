package coinbase

import (
	"github.com/johnmillner/robo-macd/internal/yaml"
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
	err := yaml.ParseYaml("../configs/coinbase.yaml", &coinbase)
	return coinbase, err
}
