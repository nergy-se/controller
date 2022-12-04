package config

// SEK default for now.
type Currency string

// highlow25|25-day
// highlow25|25-12h (this is what i run today in stampzilla-tibber).
// avg85|115-24h.
type LevelFormula string

// currently thermiagenesis is tested and supported.
type HeatControlType string

var HeatControlTypeThermiaGenesis = HeatControlType("thermiagenesis")

type Config struct {
	HeatControlType HeatControlType `json:"heatControlType"`
	// Always consider price sheep if below this level
	ConsideredCheap float64 `json:"consideredCheap"`

	// HotWaterHours how many hours during the cheapest period of one day we should prioritize hotwater.
	HotWaterHours int64 `json:"hotWaterHours"`

	// How to calculate the price level
	LevelFormula LevelFormula `json:"levelFormula"`
	Currency     Currency     `json:"currency"`

	// if not 0.0 we need to calculate if fjärrvärme is more ecomoic for that (day or hour?? TODO)
	DistrictHeatingPrice float64 `json:"districtHeatingPrice"`
}
