package coinbase

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Coinbase struct {
	Auth struct {
		Key    string
		Secret string
	}
}

// grabs coinbase auth details from configs/coinbaseAuth.yml
func getCredentials(path string) Coinbase {
	authFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("could not read configs/coinbase.yml - please make sure to populate")
	}

	coinbase := Coinbase{}
	err = yaml.Unmarshal(authFile, &coinbase)
	if err != nil {
		log.Fatalln("could not deserialize configs/coinbase.yml - please make sure to populate")
	}

	return coinbase
}
