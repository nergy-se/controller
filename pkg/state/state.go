package state

const (
	ValveStateHeating   ValveState = false
	ValveStateWarmWater ValveState = true
)

type ValveState bool

type State struct {
	Indoor             *float64
	Outdoor            *float64
	HeatCarrierForward *float64
	HeatCarrierReturn  *float64
	RadiatorForward    *float64
	RadiatorReturn     *float64 // FINNS DENNA ENS?
	BrineIn            *float64
	BrineOut           *float64
	HotGasCompressor   *float64
	WarmWater          *float64
	Compressor         *bool
	Alarm              *bool
	SwitchValve        *bool
	PumpBrine          *bool
	PumpHeat           *bool
	PumpRadiator       *bool
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
		m["compressor"] = boolToInt(*s.Compressor)
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
