# Ichor

Ichor is a Discord bot built with [arikawa](https://github.com/diamondburned/arikawa) that provides management tools for managing type 1 diabetes inspired by the likes of [Nightscout](http://www.nightscout.info/) and [LoopKit](https://loopkit.github.io/loopdocs/). It is intended to be used in tandem with the **Dexcom G6** CGM (Continuous Glucose Monitor) system. 

The server reads data from Dexcom via the Share API and stores it locally on a [boltDB](https://github.com/boltdb/bolt) instance. The data is then displayed graphically and blood glucose values are forecasted 30-minutes ahead using a LSTM model.

This project is **highly** experimental and is **not** intended to be used for therapy.

## Features 

Currently, Ichor offers a barebones set of commands

* `/glucose` fetches glucose observations from the last 12h and makes glucose forecasts up to 6 hours ahead, by feeding predictions into the same model. It also generates an accompanying chart.
![dailyOverview](docs/media/dailyOverlay.png)
The chart also displays any registered insulin and carbohydrate intake within the time period.
![dailyOverviewPlot](docs/media/dailyOverlayPlot.png)
* `/weekly` generates an weekly overview of glucose values. This includes the proportion of time spent in range, below range, above range, and the overall change since last week.
![weeklyOverview](docs/media/weeklyOverlay.png)
![dailyOverviewPlot](docs/media/weeklyOverlayPlot.png)
* `/insulin` registers the given insulin intake. Currently only supports `rapid` (insulin lispro) and `long` (insulin degludec) insulin types.
* `/carbohydrate` registers the given carbohydrate intake. Currently does not include information on the glycemic index.

## Setup

Will be included once the project reaches a more stable stage.

## Details

A slightly more detailed overview of the project.

* A timeseries abstraction is built over [bolt](https://github.com/boltdb/bolt) to more easily store timeseries on the embedded database. For the described use cases, performance is not critical.
* A functional Dexcom client is also available that makes use of the more obscure Share API to fetch glucose + trend data in real-time.
* A neural network was trained to predict future glucose values based on past glucose values, carbohydrate and insulin intake. This is very experimental, and is more of a foray into Machine Learning. The training set includes roughly 1 month of data.

## To-Dos

* Improve the user experience via Discord using an "always-on" mode rather than the current `/glucose` and `/weekly` method.
* Setup proper model serving instead of using Docker images and tensorflow.
* Various code refactoring and optimization. Also unit tests and more logging.