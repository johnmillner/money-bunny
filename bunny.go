package main

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/johnmillner/money-bunny/config"
	"github.com/johnmillner/money-bunny/internal"
	"github.com/johnmillner/money-bunny/io"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"runtime/debug"
	"sync"
	"time"
)

func main() {
	// read in configs
	config.Config("config")

	logrus.Info("starting money bunny")

	recovery(time.Now(), func() {
		a := io.NewAlpaca()
		p := io.InitPolygon()

		today, opens, closes := a.GetMarketTime()

		if !today {
			logrus.Info("market is not open today, exiting for the day")
			return
		}

		in := opens.Add(time.Duration(viper.GetInt("trade-after-open-min")) * time.Minute)
		out := closes.Add(-1 * time.Duration(viper.GetInt("close-before-close-min")) * time.Minute)

		if time.Now().Before(in) {
			logrus.Warnf("market has not opened for today yet, waiting until %s", in.String())
			time.Sleep(in.Sub(time.Now()))
		}

		overseer := internal.InitOverseer(a, p, out)

		go func() {
			for status := range p.Statuses {
				logrus.Debug(status)
			}
		}()

		// find list of all US, marginable, easy-to-trade stocks
		symbols := internal.FilterByTradable(a)
		// further filter by ensure at least small-cap and above
		symbols = internal.FilterByCap(symbols...)

		logrus.Debugf("examining %d stocks", len(symbols))
		stocks := GetStocks(a, symbols...)
		for _, stock := range stocks {
			go overseer.Manage(stock)
		}

		time.Sleep(out.Add(time.Minute).Sub(time.Now()))
		logrus.Info("market has closed for today, exiting for the day")
	})
}

func recovery(start time.Time, f func()) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("recovering from panic %v", err)
			debug.PrintStack()
			if start.Add(time.Duration(viper.GetInt("recover-frequency-min")) * time.Minute).After(time.Now()) {
				logrus.Panicf("too many panics - will not recover due to %v", err)
				return
			}

			go recovery(time.Now(), f)
		}
	}()

	f()
}

func GetStocks(a *io.Alpaca, symbols ...string) []*internal.Stock {
	stocks := make([]*internal.Stock, 0)

	limit := viper.GetInt("trend") + viper.GetInt("snapshot-lookback-min") + 2
	chunks := SplitList(symbols, viper.GetInt("chunk-size"))

	m := sync.RWMutex{}
	wg := sync.WaitGroup{}

	start := time.Now()

	for _, chunk := range chunks {
		wg.Add(1)

		go func(chunk []string) {
			defer wg.Done()

			bars, err := a.Client.ListBars(chunk, alpaca.ListBarParams{
				Timeframe: "1Min",
				Limit:     &limit,
			})

			if err != nil {
				logrus.
					WithError(err).
					Panic("could not gather historical prices")
			}

			for symbol, bar := range bars {
				if len(bar) < limit {
					continue
				}

				if time.Now().Sub(bar[len(bar)-1].GetTime()) > 2*time.Minute {
					continue
				}

				m.Lock()
				stocks = append(stocks, internal.NewStock(symbol, bar))
				m.Unlock()
			}
		}(chunk)
	}

	wg.Wait()

	logrus.Debugf("it took %s to gather historical values for %d symbols", time.Now().Sub(start).String(), len(symbols))

	return stocks
}

func SplitList(symbols []string, chunkSize int) [][]string {
	chunks := make([][]string, 0)

	for i := 0; i < len(symbols); i += chunkSize {
		stop := i + chunkSize
		if len(symbols) < stop {
			stop = len(symbols)
		}
		chunks = append(chunks, symbols[i:stop])
	}

	return chunks
}
