package models

type Category string

const (
	CategoryDefault Category = ""
	CategorySpot    Category = "spot"
	CategoryLinear  Category = "linear"
	CategoryInverse Category = "inverse"
)

type Interval string

const (
	Interval1Min   Interval = "1"
	Interval3Min   Interval = "3"
	Interval5Min   Interval = "5"
	Interval15Min  Interval = "15"
	Interval30Min  Interval = "30"
	Interval60Min  Interval = "60"
	Interval120Min Interval = "120"
	Interval240Min Interval = "240"
	Interval360Min Interval = "360"
	Interval720Min Interval = "720"
	Interval1Day   Interval = "D"
	Interval1Week  Interval = "W"
	Interval1Month Interval = "M"
)
