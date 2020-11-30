![Go](https://github.com/johnmillner/money-bunny/workflows/Go/badge.svg)


# Code/Testing In Progress - Don't trust your money with this yet :) 

# Money Bunny
This is a day trading stock scanner and robo-trader that searches all US stocks and attempts to find stocks with a 
good risk-reward ratio and that are at the start of an upward trend. 

# Configuration
in the config.yml add your alpaca key, secret
Alternatively, these values can be stored as environment variables at 
- BUNNY_ALPACA.KEY
- BUNNY_ALPACA.SECRET

additional configurations can be found inside of config/config.yml to control risk, reward, and so on
equally these can be overriden by environment variables with the prefix `BUNNY_`

## Customizing Configuration Values
The configuration values already supplied have been tested to be profitable in paper trading accounts.

If wanting to change those values to meet your own preferences, please ensure to trial them in a paper trading account. 

Funny things can always happen when messing with values, so test to ensure that you dont blow up your account!

# Running 
Pre-Requisites:
- an account with [Alpaca](https://alpaca.markets/)
- an account with at least 25k in it to survive [PDT rule](https://www.investopedia.com/terms/p/patterndaytrader.asp)
- add your credentials to the program either through config/config.yml or through environment variables

clone or download the package - and run `go run bunny.go`

# Alpaca
[Alpaca](https://alpaca.markets/) is an API first, 0 commission broker that is used by this robo-trader to interact with equity markets

# Strategy
## Buy
- find all stocks
- filter out the stocks that are outside of our wanted risk by using ATR indicator
    - ```go
        func meetsRiskGoal(stock *stock.Stock) bool {
        	tradeRisk := viper.GetFloat64("stop-loss-atr-ratio") * stock.Atr[len(stock.Atr)-1] / stock.Price.Peek()
        	upperRisk := viper.GetFloat64("risk") * (1 + viper.GetFloat64("exposure-tolerance"))
        	lowerRisk := viper.GetFloat64("risk") * (1 - viper.GetFloat64("exposure-tolerance"))
        
        	return tradeRisk > lowerRisk && tradeRisk < upperRisk
        }
        ```
- filter out the stocks that do not have an entry signal in the past minute using MACD and Trend indicator
    -  ```go
       func (s *Stock) IsReadyToBuy() bool {
       	return s.IsBelowTrend() && s.IsUpwardsTrend() && s.IsBuyingMacdCrossOver()
       }
        ``` 
- buy those stocks calculating stopLoss and takeProfit from the ATR and configured risk parameters
    - ```go
        func getOrderParameters(s stock.Stock, a *io.Alpaca, budget float64) (float64, float64, float64, float64, float64) {
        	quote := a.GetQuote(s.Symbol)
        	exposure := budget * viper.GetFloat64("risk")
        	price := float64(quote.Last.AskPrice - (quote.Last.AskPrice-quote.Last.BidPrice)/2)      
        	
        	tradeRisk := viper.GetFloat64("stop-loss-atr-ratio") * s.Atr[len(s.Atr)-1]
        	rewardToRisk := viper.GetFloat64("risk-reward")
        	stopLossMax := viper.GetFloat64("stop-loss-max")
        
        	takeProfit := price + (rewardToRisk * tradeRisk)
        	stopLoss := price - tradeRisk
        	stopLimit := price - (1+stopLossMax)*tradeRisk
        
        	qty := math.Round(exposure / tradeRisk)
        
        	//ensure we dont go over
        	for qty * price > budget {
        		qty = qty - 1
        	}
        
        	return price, qty, takeProfit, stopLoss, stopLimit
        }
        ```
 
## Sell
- sell all stocks if the market is close to closing (2m)
- sell all stocks if they have not closed within a defined time (30m)
- sell a stock if it has meet its take-profit or stop-loss (completing automatically with bracket orders during buy)
- sell a stock if the MACD and Trend indicators show an exit indication
    - ```go
        func (s *Stock) IsReadyToSell() bool {
            return !s.IsBelowTrend() && s.IsDownwardsTrend() && s.IsSellingMacdCrossUnder()
        }
        ```

# Logging
while the program is running, the console will log all trades, including symbol, price, take profit, stop loss, stop limit, and qty
```
2020/11/29 02:34:08 buying NHC:
	total:      22332.509254
	qty:        349.000000
	maxProfit:  375.502682
	maxLoss:    250.33512142059453
	price:      63.989998
	takeProfit: 65.065937
	stopLoss:   63.272705
```
inside of snapshots directory there will be html pages showing the graphs of the stock bought which it based its decision on

they will be named with the time it was purchased and the stock given

![graphs html showing stock snapshot during buy](example/graphExample.png?raw=true "graphs html showing stock snapshot during buy")

# Disclaimer
As with all things, both common sense and within the GPL-3.0 License agreed to on use of this program, 
there is no guarantee, warranty, or liability agreed or expressed when using this program. 

While this program is made in good faith to be profitable and bug-free - there is no promise of profits.
There is however, risk. Risk of losing money. Risk of margin-call. Risk of buggy software.

This program is designed to day-trade any stock that meets certain mathematical technical indicators. 
There is no advisory involved, and it views equities as a collection of potentially flawed numbers, 
and its decisions as a series of potentially flawed equations. 
It is important to understand that there is significant risk in 
- market instability
- day trading
- use of margin
- relying heavily on technical indicators
- code flaws
- data flaws

This program also uses margin by default - if unfamiliar with that risk - 
or if not wanting that additional risk, please ensure to set "margin-multiplier" in the config to 1.00 or less 

As a general rule of thumb: do not trade with money you cannot lose.
As a recomendation - read through this code yourself to understand what it does and how it works. 

If you do see something that could be improved or safe-guarded against - please open up an incident or a PR and let's fix it! 