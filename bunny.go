package main

import (
	"github.com/johnmillner/money-bunny/config"
	"github.com/johnmillner/money-bunny/internal"
	"github.com/spf13/viper"
	"log"
	"runtime/debug"
	"time"
)

func main() {
	log.Print("starting money bunny")

	// read in configs
	config.Config("config")

	recovery(time.Now(), func() {
		a := internal.NewAlpaca()
		p := internal.InitPolygon()

		today, opens, closes := a.GetMarketTime()

		if !today {
			log.Printf("market is not open today, exiting for the day")
			return
		}

		in := opens.Add(time.Duration(viper.GetInt("trade-after-open-min")) * time.Minute)
		out := closes.Add(-1 * time.Duration(viper.GetInt("close-before-close-min")) * time.Minute)

		if time.Now().Before(in) {
			log.Printf("market has not opened for today yet, waiting until %s", in.String())
			time.Sleep(in.Sub(time.Now()))
		}

		overseer := internal.InitOverseer(a, p, out)
		// find list of all US, marginable, easy-to-trade stocks
		symbols := internal.FilterByTradability(a)
		// further filter by ensure at least small-cap and above
		symbols = internal.FilterByCap(symbols...)

		stocks := a.GetStocks(symbols...)
		for _, stock := range stocks {
			go overseer.Manage(stock)
		}

		time.Sleep(out.Add(time.Minute).Sub(time.Now()))
		log.Printf("market has closed for today, exiting for the day")
	})
}

func recovery(start time.Time, f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from panic %v", err)
			debug.PrintStack()
			if start.Add(time.Duration(viper.GetInt("recover-frequency-min")) * time.Minute).After(time.Now()) {
				log.Panicf("too many panics - will not recover due to %v", err)
				return
			}

			go recovery(time.Now(), f)
		}
	}()

	f()
}
