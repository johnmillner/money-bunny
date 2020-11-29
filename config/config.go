package config

import (
	"github.com/spf13/viper"
	"log"
	"strings"
)

func Config() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	viper.EnvKeyReplacer(strings.NewReplacer("_", "."))
	viper.SetEnvPrefix("BUNNY")
	viper.AutomaticEnv()
	viper.WatchConfig()

	err := viper.ReadInConfig()

	if err != nil {
		log.Panicf("Fatal error config file: %s", err)
	}
}
