package dummy

import (
	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

type Dummy struct {
}

func New() *Dummy {
	return &Dummy{}
}
func Pointer[K any](val K) *K {
	return &val
}

func (ts *Dummy) State() (*state.State, error) {
	s := &state.State{
		Indoor:             Pointer(21.1),
		Outdoor:            Pointer(11.1),
		HeatCarrierForward: nil,
		HeatCarrierReturn:  nil,
		RadiatorForward:    nil,
		RadiatorReturn:     nil,
		BrineIn:            Pointer(4.2),
		BrineOut:           Pointer(2.2),
		HotGasCompressor:   nil,
		WarmWater:          nil,
		Compressor:         Pointer(true),
		Alarm:              nil,
		SwitchValve:        nil,
		PumpBrine:          nil,
		PumpHeat:           nil,
		PumpRadiator:       nil,
	}

	return s, nil
}
func (ts *Dummy) BlockHeating() error {
	logrus.Info("dummy: BlockHeating")
	return nil
}
func (ts *Dummy) AllowHeating() error {
	logrus.Info("dummy: AllowHeating")
	return nil
}

func (ts *Dummy) BlockHotwater() error {
	logrus.Info("dummy: BlockHotwater")
	return nil
}
func (ts *Dummy) AllowHotwater() error {
	logrus.Info("dummy: AllowHotwater")
	return nil
}

func (ts *Dummy) BoostHotwater() error {
	// TODO
	logrus.Info("dummy: BoostHotwater")

	return nil

}
