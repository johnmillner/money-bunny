package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

func Config(path string) {

	// setup configs
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.EnvKeyReplacer(strings.NewReplacer("_", "."))
	viper.SetEnvPrefix("BUNNY")
	viper.AutomaticEnv()

	f, err := os.OpenFile(
		fmt.Sprintf("logs/%s.log", time.Now().Format("2006-01-02")),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0755)

	if err != nil {
		logrus.
			WithError(err).
			Panic("could not open file to store logs")
	}

	logrus.SetOutput(&fileAndConsole{
		console: os.Stderr,
		file:    f,
	})

	viper.WatchConfig()
	err = viper.ReadInConfig()
	if err != nil {
		logrus.WithError(err).Panic("Fatal error config file")
	}

	updateLevelAndFormat()

	viper.OnConfigChange(func(e fsnotify.Event) {
		updateLevelAndFormat()
		logrus.Debugf("Config file updated: %s", e.Name)
	})
}

func updateLevelAndFormat() {
	logrus.SetLevel(getLogLevel(viper.GetString("log-level")))
	if viper.GetBool("log-json") {
		logrus.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: true,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}

type fileAndConsole struct {
	console, file *os.File
}

func (w fileAndConsole) Write(p []byte) (n int, err error) {
	_, _ = w.console.Write(p)
	return w.file.Write(p)
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
