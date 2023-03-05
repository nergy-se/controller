package state

const (
	ValveStateHeating   ValveState = false
	ValveStateWarmWater ValveState = true
)

type ValveState bool

type State struct {
	Indoor             *float64 `json:"indoor,omitempty"`
	Outdoor            *float64 `json:"outdoor,omitempty"`
	HeatCarrierForward *float64 `json:"heatCarrierForward,omitempty"`
	HeatCarrierReturn  *float64 `json:"heatCarrierReturn,omitempty"`
	RadiatorForward    *float64 `json:"radiatorForward,omitempty"`
	RadiatorReturn     *float64 `json:"radiatorReturn,omitempty"` // FINNS DENNA ENS?
	BrineIn            *float64 `json:"brineIn,omitempty"`
	BrineOut           *float64 `json:"brineOut,omitempty"`
	HotGasCompressor   *float64 `json:"hotGasCompressor,omitempty"`
	WarmWater          *float64 `json:"warmWater,omitempty"`
	Compressor         *float64 `json:"compressor,omitempty"`
	Alarm              *bool    `json:"alarm,omitempty"`
	SwitchValve        *bool    `json:"switchValve,omitempty"`
	PumpBrine          *bool    `json:"pumpBrine,omitempty"`
	PumpHeat           *bool    `json:"pumpHeat,omitempty"`
	PumpRadiator       *bool    `json:"pumpRadiator,omitempty"`
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
		m["pumpBrine"] = boolToInt(*s.PumpBrine)
	}
	if s.PumpHeat != nil {
		m["pumpHeat"] = boolToInt(*s.PumpHeat)
	}
	if s.PumpRadiator != nil {
		m["pumpRadiator"] = boolToInt(*s.PumpRadiator)
	}

	return m
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
