package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"strings"
)

func Config(path string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.EnvKeyReplacer(strings.NewReplacer("_", "."))
	viper.SetEnvPrefix("BUNNY")
	viper.AutomaticEnv()
	viper.WatchConfig()

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file updated: %s", e.Name)
	})

	err := viper.ReadInConfig()

	if err != nil {
		log.Panicf("Fatal error config file: %s", err)
	}
}
