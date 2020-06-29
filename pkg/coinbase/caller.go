package coinbase

type Coinbase struct {
	Auth struct {
		Key    string
		Secret string
	}
}

func getCoinbase() (Coinbase, error) {
	coinbase := Coinbase{}
	err := GetYamlConfig("coinbase.yaml", &coinbase)
	return coinbase, err
}
