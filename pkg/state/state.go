package state

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
