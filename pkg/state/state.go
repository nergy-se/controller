package state

import "time"

const (
	ValveStateHeating   ValveState = false
	ValveStateWarmWater ValveState = true
)

type ValveState bool

type State struct {
	Time                     time.Time `json:"time"`
	Indoor                   *float64  `json:"indoor,omitempty"`
	IndoorSetpoint           *float64  `json:"indoorSetpoint,omitempty"`
	Outdoor                  *float64  `json:"outdoor,omitempty"`
	HeatCarrierForward       *float64  `json:"heatCarrierForward,omitempty"`
	HeatCarrierReturn        *float64  `json:"heatCarrierReturn,omitempty"`
	RadiatorForward          *float64  `json:"radiatorForward,omitempty"`
	RadiatorReturn           *float64  `json:"radiatorReturn,omitempty"` // FINNS DENNA ENS?
	BrineIn                  *float64  `json:"brineIn,omitempty"`
	BrineOut                 *float64  `json:"brineOut,omitempty"`
	HotGasCompressor         *float64  `json:"hotGasCompressor,omitempty"`
	WarmWater                *float64  `json:"warmWater,omitempty"`
	Compressor               *float64  `json:"compressor,omitempty"`
	Alarm                    *bool     `json:"alarm,omitempty"`
	SwitchValve              *bool     `json:"switchValve,omitempty"`
	PumpBrine                *float64  `json:"pumpBrine,omitempty"`
	PumpHeat                 *float64  `json:"pumpHeat,omitempty"`
	PumpRadiator             *float64  `json:"pumpRadiator,omitempty"`
	SuperHeatTemperature     *float64  `json:"superHeatTemperature,omitempty"`
	SuctionGasTemperature    *float64  `json:"suctionGasTemperature,omitempty"`
	LowPressureSidePressure  *float64  `json:"lowPressureSidePressure,omitempty"`
	HighPressureSidePressure *float64  `json:"highPressureSidePressure,omitempty"`
	COP                      *float64  `json:"cop,omitempty"`

	HeatingAllowed  *bool `json:"heatingAllowed,omitempty"`
	HotwaterAllowed *bool `json:"hotwaterAllowed,omitempty"`
}
