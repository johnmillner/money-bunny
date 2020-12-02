package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

func Config(path string) {
	// set up logs
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	// setup configs
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.EnvKeyReplacer(strings.NewReplacer("_", "."))
	viper.SetEnvPrefix("BUNNY")
	viper.AutomaticEnv()
	viper.WatchConfig()

	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.SetLevel(getLogLevel(viper.GetString("log-level")))
		logrus.Debugf("Config file updated: %s", e.Name)
	})

	err := viper.ReadInConfig()

	if err != nil {
		logrus.WithError(err).Panic("Fatal error config file")
	}
}

func getLogLevel(level string) logrus.Level {
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}
