package dummy

import (
	"math/rand"

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
	compressor := float64(rand.Intn(100-20) + 20)
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
		Compressor:         &compressor,
		Alarm:              nil,
		SwitchValve:        nil,
		PumpBrine:          nil,
		PumpHeat:           nil,
		PumpRadiator:       nil,
	}

	return s, nil
}

func (ts *Dummy) AllowHeating(b bool) error {
	logrus.Info("dummy: AllowHeating: ", b)
	return nil
}

func (ts *Dummy) AllowHotwater(b bool) error {
	logrus.Info("dummy: AllowHotwater: ", b)
	return nil
}

func (ts *Dummy) BoostHotwater(b bool) error {
	// TODO
	logrus.Info("dummy: BoostHotwater: ", b)

	return nil

}
