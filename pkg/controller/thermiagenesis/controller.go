package thermiagenesis

import (
	"fmt"

	"github.com/goburrow/modbus"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/state"
)

type Thermiagenesis struct {
	client modbus.Client
}

func New(client modbus.Client) *Thermiagenesis {
	return &Thermiagenesis{
		client: client,
	}
}

func (ts *Thermiagenesis) State() (*state.State, error) {
	s := &state.State{
		Indoor:             nil,
		Outdoor:            nil,
		HeatCarrierForward: nil,
		HeatCarrierReturn:  nil,
		RadiatorForward:    nil,
		RadiatorReturn:     nil,
		BrineIn:            nil,
		BrineOut:           nil,
		HotGasCompressor:   nil,
		WarmWater:          nil,
		Compressor:         nil,
		Alarm:              nil,
		SwitchValve:        nil,
		PumpBrine:          nil,
		PumpHeat:           nil,
		PumpRadiator:       nil,
	}
	var err error

	s.BrineIn, err = controller.Scale100itof(ts.readInputRegister(10)) // 10 brine in scale 100
	if err != nil {
		return s, err
	}

	s.BrineOut, err = controller.Scale100itof(ts.readInputRegister(11)) // 11 brine out scale 100
	if err != nil {
		return s, err
	}
	s.Outdoor, err = controller.Scale100itof(ts.readInputRegister(13)) // 13 Outdoor temp scale 100
	if err != nil {
		return s, err
	}
	s.Indoor, err = controller.Scale10itof(ts.readInputRegister(121)) // Room temperature sensor scale 10
	if err != nil {
		return s, err
	}

	s.WarmWater, err = controller.Scale100itof(ts.readInputRegister(17)) // 17 tank Tap water weighted temperature scale 100
	if err != nil {
		return s, err
	}
	s.Compressor, err = controller.Scale100itof(ts.readInputRegister(54)) // Compressor speed percent scale 100
	if err != nil {
		return s, err
	}
	s.Indoor, err = controller.Scale100itof(ts.readInputRegister(121)) // Room temperature sensor scale 10
	if err != nil {
		return s, err
	}
	s.RadiatorForward, err = controller.Scale100itof(ts.readInputRegister(12)) // System supply line temperature scale 100 vad är detta?! visar bara 200.0 så inte inkopplad?
	// https://github.com/CJNE/thermiagenesis/issues/157#issuecomment-1250896092
	if err != nil {
		return s, err
	}
	s.RadiatorReturn, err = controller.Scale100itof(ts.readInputRegister(27)) // input reg 27 System return line temperature. visar 0 hos per
	if err != nil {
		return s, err
	}

	s.HeatCarrierForward, err = controller.Scale100itof(ts.readInputRegister(9)) // input reg 9 Condenser out temperature som visar 47.26 kan vara detta? HeatCarrierForward!
	if err != nil {
		return s, err
	}
	s.HeatCarrierReturn, err = controller.Scale100itof(ts.readInputRegister(8)) // input reg 8 Condenser in som visar 42.08 kan vara detta? HeatCarrierReturn!
	if err != nil {
		return s, err
	}
	s.PumpBrine, err = controller.Scale100itof(ts.readInputRegister(44)) // input reg 44 Brine circulation pump speed (%) just nu 66.81 PumpBrine
	if err != nil {
		return s, err
	}
	s.PumpHeat, err = controller.Scale100itof(ts.readInputRegister(39)) // input reg 39 Condenser circulation pump speed (%) just nu 60.1 PumpHeat
	if err != nil {
		return s, err
	}

	// input reg 147 Desired temperature distribution circuit Mix valve 1 verkar vara nuvarande uträknade börvärde? tex 38.08
	// input reg 7 Discharge pipe temperature verkar vara hetgasen? just nu 81.12
	// input reg 1 Currently running: First prioritised demand *1
	//  1: Manual operation, 2: Defrost, 3: Hot water, 4: Heat, 5: Cool, 6: Pool, 7: Anti legionella, 98: Standby 99: No demand 100: OFF
	// input reg 18 System supply line calculated set point just nu 47.18

	return s, nil
}

func (ts *Thermiagenesis) AllowHeating(b bool) error {
	return nil
}

func (ts *Thermiagenesis) AllowHotwater(b bool) error {
	return nil
}

func (ts *Thermiagenesis) BoostHotwater(b bool) error {
	start := 41
	stop := 50
	if b {
		start = 50
		stop = 55
	}
	_, err := ts.client.WriteSingleRegister(22, uint16(start*100)) // 100 = 1c
	if err != nil {
		return fmt.Errorf("error writeTemps 22: %w", err)
	}

	_, err = ts.client.WriteSingleRegister(23, uint16(stop*100))
	if err != nil {
		return fmt.Errorf("error writeTemps 23: %w", err)
	}
	return nil

}

func (ts *Thermiagenesis) readInputRegister(address uint16) (int, error) {
	b, err := ts.client.ReadInputRegisters(address, 1)
	return controller.Decode(b), err
}

func (ts *Thermiagenesis) readHoldingRegister(address uint16) (int, error) {
	b, err := ts.client.ReadHoldingRegisters(address, 1)
	return controller.Decode(b), err
}
