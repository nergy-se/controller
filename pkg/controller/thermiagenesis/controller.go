package thermiagenesis

import (
	"fmt"

	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/modbusclient"
	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

type Thermiagenesis struct {
	client   modbusclient.Client
	readonly bool
}

func New(client modbusclient.Client, readonly bool) *Thermiagenesis {
	return &Thermiagenesis{
		client:   client,
		readonly: readonly,
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

	s.BrineIn, err = controller.Scale100itof(ts.client.ReadInputRegister(10)) // 10 brine in scale 100
	if err != nil {
		return s, err
	}

	s.BrineOut, err = controller.Scale100itof(ts.client.ReadInputRegister(11)) // 11 brine out scale 100
	if err != nil {
		return s, err
	}
	s.Outdoor, err = controller.Scale100itof(ts.client.ReadInputRegister(13)) // 13 Outdoor temp scale 100
	if err != nil {
		return s, err
	}
	s.Indoor, err = controller.Scale10itof(ts.client.ReadInputRegister(121)) // Room temperature sensor scale 10
	if err != nil {
		return s, err
	}

	s.WarmWater, err = controller.Scale100itof(ts.client.ReadInputRegister(17)) // 17 tank Tap water weighted temperature scale 100
	if err != nil {
		return s, err
	}
	s.Compressor, err = controller.Scale100itof(ts.client.ReadInputRegister(54)) // Compressor speed percent scale 100
	if err != nil {
		return s, err
	}
	s.Indoor, err = controller.Scale100itof(ts.client.ReadInputRegister(121)) // Room temperature sensor scale 10
	if err != nil {
		return s, err
	}
	s.RadiatorForward, err = controller.Scale100itof(ts.client.ReadInputRegister(12)) // System supply line temperature scale 100 vad är detta?! visar bara 200.0 så inte inkopplad?
	// https://github.com/CJNE/thermiagenesis/issues/157#issuecomment-1250896092
	if err != nil {
		return s, err
	}
	s.RadiatorReturn, err = controller.Scale100itof(ts.client.ReadInputRegister(27)) // input reg 27 System return line temperature. visar 0 hos per
	if err != nil {
		return s, err
	}

	s.HeatCarrierForward, err = controller.Scale100itof(ts.client.ReadInputRegister(9)) // input reg 9 Condenser out temperature som visar 47.26 kan vara detta? HeatCarrierForward!
	if err != nil {
		return s, err
	}
	s.HeatCarrierReturn, err = controller.Scale100itof(ts.client.ReadInputRegister(8)) // input reg 8 Condenser in som visar 42.08 kan vara detta? HeatCarrierReturn!
	if err != nil {
		return s, err
	}
	s.PumpBrine, err = controller.Scale100itof(ts.client.ReadInputRegister(44)) // input reg 44 Brine circulation pump speed (%) just nu 66.81 PumpBrine
	if err != nil {
		return s, err
	}
	s.PumpHeat, err = controller.Scale100itof(ts.client.ReadInputRegister(39)) // input reg 39 Condenser circulation pump speed (%) just nu 60.1 PumpHeat
	if err != nil {
		return s, err
	}

	s.HotGasCompressor, err = controller.Scale100itof(ts.client.ReadInputRegister(7)) // input reg 7 Discharge pipe temperature
	if err != nil {
		return s, err
	}
	s.SuperHeatTemperature, err = controller.Scale100itof(ts.client.ReadInputRegister(125)) // input reg 125 Superheat temperature
	if err != nil {
		return s, err
	}
	s.SuctionGasTemperature, err = controller.Scale100itof(ts.client.ReadInputRegister(130)) // input reg 130 Suction gas temperature
	if err != nil {
		return s, err
	}
	s.LowPressureSidePressure, err = controller.Scale100itof(ts.client.ReadInputRegister(127)) // input reg 127 Low pressure side, pressure (bar(g))
	if err != nil {
		return s, err
	}
	s.HighPressureSidePressure, err = controller.Scale100itof(ts.client.ReadInputRegister(128)) // input reg 128 High pressure side, pressure (bar(g))
	if err != nil {
		return s, err
	}

	// input reg 147 Desired temperature distribution circuit Mix valve 1 verkar vara nuvarande uträknade börvärde? tex 38.08
	// input reg 1 Currently running: First prioritised demand *1
	//  1: Manual operation, 2: Defrost, 3: Hot water, 4: Heat, 5: Cool, 6: Pool, 7: Anti legionella, 98: Standby 99: No demand 100: OFF
	// input reg 18 System supply line calculated set point just nu 47.18
	// write single coil(5) enable heat 9
	// write single coil(5) enable tap water 8

	// operational mode WriteSingleRegister 0 1: OFF, 2: Standby, 3: ON/Auto

	return s, nil
}

func (ts *Thermiagenesis) AllowHeating(b bool) error {
	_, err := ts.client.WriteSingleRegister(9, modbusclient.CoilValue(b))
	return err
}

func (ts *Thermiagenesis) AllowHotwater(b bool) error {
	_, err := ts.client.WriteSingleRegister(8, modbusclient.CoilValue(b))
	return err
}

func (ts *Thermiagenesis) BoostHotwater(b bool) error {
	start := 41
	stop := 50
	if b {
		start = 50
		stop = 55
	}

	logrus.WithFields(logrus.Fields{"start": start, "stop": stop}).Debugf("thermiagenesis: boosthotwater")
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
