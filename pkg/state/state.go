package state

const (
	ValveStateHeating   ValveState = false
	ValveStateWarmWater ValveState = true
)

type ValveState bool

type State struct {
	Indoor                   *float64 `json:"indoor,omitempty"`
	Outdoor                  *float64 `json:"outdoor,omitempty"`
	HeatCarrierForward       *float64 `json:"heatCarrierForward,omitempty"`
	HeatCarrierReturn        *float64 `json:"heatCarrierReturn,omitempty"`
	RadiatorForward          *float64 `json:"radiatorForward,omitempty"`
	RadiatorReturn           *float64 `json:"radiatorReturn,omitempty"` // FINNS DENNA ENS?
	BrineIn                  *float64 `json:"brineIn,omitempty"`
	BrineOut                 *float64 `json:"brineOut,omitempty"`
	HotGasCompressor         *float64 `json:"hotGasCompressor,omitempty"`
	WarmWater                *float64 `json:"warmWater,omitempty"`
	Compressor               *float64 `json:"compressor,omitempty"`
	Alarm                    *bool    `json:"alarm,omitempty"`
	SwitchValve              *bool    `json:"switchValve,omitempty"`
	PumpBrine                *float64 `json:"pumpBrine,omitempty"`
	PumpHeat                 *float64 `json:"pumpHeat,omitempty"`
	PumpRadiator             *float64 `json:"pumpRadiator,omitempty"`
	SuperHeatTemperature     *float64 `json:"superHeatTemperature,omitempty"`
	SuctionGasTemperature    *float64 `json:"suctionGasTemperature,omitempty"`
	LowPressureSidePressure  *float64 `json:"lowPressureSidePressure,omitempty"`
	HighPressureSidePressure *float64 `json:"highPressureSidePressure,omitempty"`
	COP                      *float64 `json:"cop,omitempty"`

	HeatingAllowed  *bool `json:"heatingAllowed,omitempty"`
	HotwaterAllowed *bool `json:"hotwaterAllowed,omitempty"`
}

func (s State) Map() map[string]interface{} {
	m := make(map[string]interface{})
	if s.Indoor != nil {
		m["indoor"] = *s.Indoor
	}
	if s.Outdoor != nil {
		m["outdoor"] = *s.Outdoor
	}
	if s.HeatCarrierForward != nil {
		m["heatCarrierForward"] = *s.HeatCarrierForward
	}
	if s.HeatCarrierReturn != nil {
		m["heatCarrierReturn"] = *s.HeatCarrierReturn
	}
	if s.RadiatorForward != nil {
		m["radiatorForward"] = *s.RadiatorForward
	}
	if s.RadiatorReturn != nil {
		m["radiatorReturn"] = *s.RadiatorReturn
	}
	if s.BrineIn != nil {
		m["brineIn"] = *s.BrineIn
	}
	if s.BrineOut != nil {
		m["brineOut"] = *s.BrineOut
	}
	if s.HotGasCompressor != nil {
		m["hotGasCompressor"] = *s.HotGasCompressor
	}
	if s.WarmWater != nil {
		m["warmWater"] = *s.WarmWater
	}
	if s.Compressor != nil {
		m["compressor"] = *s.Compressor
	}
	if s.Alarm != nil {
		m["alarm"] = boolToInt(*s.Alarm)
	}
	if s.SwitchValve != nil {
		m["switchValve"] = boolToInt(*s.SwitchValve)
	}
	if s.PumpBrine != nil {
		m["pumpBrine"] = *s.PumpBrine
	}
	if s.PumpHeat != nil {
		m["pumpHeat"] = *s.PumpHeat
	}
	if s.PumpRadiator != nil {
		m["pumpRadiator"] = *s.PumpRadiator
	}
	if s.SuperHeatTemperature != nil {
		m["superHeatTemperature"] = *s.SuperHeatTemperature
	}
	if s.SuctionGasTemperature != nil {
		m["suctionGasTemperature"] = *s.SuctionGasTemperature
	}
	if s.LowPressureSidePressure != nil {
		m["lowPressureSidePressure"] = *s.LowPressureSidePressure
	}
	if s.HighPressureSidePressure != nil {
		m["highPressureSidePressure"] = *s.HighPressureSidePressure
	}
	if s.COP != nil {
		m["cop"] = *s.COP
	}
	if s.HeatingAllowed != nil {
		m["heatingAllowed"] = boolToInt(*s.HeatingAllowed)
	}
	if s.HotwaterAllowed != nil {
		m["hotwaterAllowed"] = boolToInt(*s.HotwaterAllowed)
	}

	return m
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
