package io

import (
	"github.com/johnmillner/money-bunny/config"
	"log"
	"testing"
)

func TestGetMarketCap(t *testing.T) {
	config.Config("../config")

	log.Printf("%f", GetMarketCap("TSLA"))
}
