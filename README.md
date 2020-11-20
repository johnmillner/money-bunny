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

# manager strategy - buying
use 200 ema as longer term true value 
 - will block buys when price is above ema
use acceleration and velocity of 200 ema as trend directional indicator
 - will block buys when acceleration or velocity of ema is <0
use macd to determine trend entry signal
 - only buy when positive crossover with histo<0 (delta(histo=0))
use 2*ATR to determine stop-loss
use 2&ATR*1.5 to determine take-profit
 
## executor strategy 
* an opportunity is noticed for a particular stock as communicated by the manager
* the configures max trade value is received
* confirmation is made with the portfolio that there is capital available to make the trade
* the stop loss is generated from the price-2*atr value
* the take profit is calculated by 
* the configured risk/reward is calculated to determine take-profit

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
 
 
