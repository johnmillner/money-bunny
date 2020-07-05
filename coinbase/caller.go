package coinbase

import (
	"log"
)

// Coinbase contains coinbase specific settings and authnetication
type Coinbase struct {
	Auth struct {
		Key    string
		Secret string
	}
}

func getCoinbase() (Coinbase, error) {
	log.Println("test")
	coinbase := Coinbase{}
	err := reader.GetYamlConfig("coinbase.yaml", &coinbase)
	return coinbase, err
}
