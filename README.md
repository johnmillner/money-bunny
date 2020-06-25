# robo-macd
Simple MACD based Robo-Advisor

# Code In Progress

Inspired by the video here by [Trading Rush](https://www.youtube.com/watch?v=nmffSjdZbWQ):

[![I risked MACD Trading Strategy 100 TIMES Here’s What Happened...](https://img.youtube.com/vi/nmffSjdZbWQ/0.jpg)](https://www.youtube.com/watch?v=nmffSjdZbWQ)

This is a MACD based robo-advisor that will buy and sell crypto from coinbase based solely on that crypto's MACD.

# configuration
## configs/coinbaseAuth.yml
this is a secure file that contains your own coinbase api authentication details. There are two values - it is on the gitignore list for safety.
```yaml
key: test
secret: test
```
you can retrieve these details by ...todo

## configs/config.yml
configuration file that details stoploss percentage, profit target, macd and trend settings, etc.
```yaml
decisionMaker:
  stopLoss: .25 # set stop loss order if price goes below 25% of original price
  profitMax: .50 # sell order if price goes above 50% above original price

macdCalculator:
  macd: # values used to calculate macd
    12-period-ema: 12
    26-period-ema: 26
    9-period-ema: 9
  trend: # values used to calculate trend line
    trend-ema-period: 200
  period: P5M #ISO 8601 format to describe period to work - default is 5 minutes
  ```

# To Use
... todo
