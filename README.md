![Go](https://github.com/johnmillner/robo-macd/workflows/Go/badge.svg)


# Code In Progress - Don't trust your money with this yet :) 

# Stock Conveyor 
Robo-Advisor that will trade stocks based off of MACD crossover

Strategy inspired by the video here by [Trading Rush](https://www.youtube.com/watch?v=nmffSjdZbWQ):

# Concept
This program will scan the stock market - determining which stocks are having a MACD Crossover event. 
When it determines that a stock is crossing over positively, it will initiate a buy order - and equally when it determines 
a held stock is crossing over negatively, it will initiate a sell order. 
Multiple time scans can be run in parallel such that there are essentially conveyor belts of opportunities available. 

There could be a conveyor belt fo 1hr period, 4hr period, 1 day, etc periods all running in parallel playing short 
term swing trades as well as longer term plays - with different stocks bought (entering the conveyor belt) and sold (exiting the conveyor belt) asynchronously

# manager strategy 
use rsi to confirm overbought or oversold momentum
 - will block buys when rsi>70%
 - will block sells when rsi<30%
use aroon indicator as a trend indicator 
 - will block buys when on a down trend
 - will block sells when on an up trend
use macd to determine trend entry and exit signals
 - only buy when positive crossover with histo<0
 - only sell when negative crossover with histo>0 

## executor strategy - buy
* an opportunity is noticed for a particular stock as communicated by the manager
* the configures max trade value is received
* confirmation is made with the portfolio that there is capital available to make the trade
* the configured wanted risk is calculated adding atr to determine the stop loss (actually risk will be greater depending on volatility)
* the configured risk/reward is calculated to determine take-profit

## executor strategy - sell
* manager notices a sell potential for a particular stock 
* this stock is checked against the portfolio to confirm if owned
* if so, close the position before the take profit/stop-loss

## notes
this algo does not take into account supports and resistances when calculating stop-losses and take-profits

# To Use
... todo

# Structure
[Diagrams found in draw.io](https://app.diagrams.net/?lightbox=1&highlight=0000ff&edit=_blank&layers=1&nav=1&title=RoboAdvisor#Uhttps%3A%2F%2Fdrive.google.com%2Fuc%3Fid%3D1fZWEaOWSyaqYmPYYk0OZuidXkcBH2hcp%26export%3Ddownload)

## Alpaca
[Alpaca](https://alpaca.markets/) is an API first, 0 commission broker that is used by this robo-advisor to interact with equity markets

## Gatherers
Gathers data from a Broker (in this case Alpaca) and transforms to a canonical representation that is then sent to a Transformer
## Transformer
Takes in the gatherer's data and transforms to wanted structure 
## Manager
Takes in the Transformed Data from the Transformers and determines for each targeted equity whether to buy/sell/hold that equity
### MACD Manager
takes in MACD Transformers chart and determines whether to buy, sell, or hold the stock at this moment
## Executor
Takes in the Managers high level decisions and translates them into actions to take place in the broker, managing the low level portfolio
## Coordinator
acts as the "main" of the program, instantiating the service and acting as a coordination hub for configurations

## UI
Provides ways to view the account, gains, trades, win/loss rate, and allows for the configuration of:
 * targeted equities and frequencies for the gathers
 * specific configurations for the gatherer(s)
 * specific configurations for the transformer(s)
 * specific configurations for the managers(s)
 * specific configurations for the executor
