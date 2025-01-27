package thermiagenesis

import (
	"fmt"

	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/api/v1/meter"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/modbusclient"
	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

type Thermiagenesis struct {
	client             modbusclient.Client
	cloudConfig        *config.CloudConfig
	readonly           bool
	calculatedCOP      float64
	heatCarrierForward float64

	heatingAllowed  bool
	hotwaterAllowed bool
}

func New(client modbusclient.Client, readonly bool, cloudConfig *config.CloudConfig) *Thermiagenesis {
	return &Thermiagenesis{
		client:        client,
		cloudConfig:   cloudConfig,
		readonly:      readonly,
		calculatedCOP: 4.0,
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
	s.RadiatorForward, err = controller.Scale100itof(ts.client.ReadInputRegister(12)) // System supply line temperature scale 100 visar bara 200.0 om inte inkopplad.
	// https://github.com/CJNE/thermiagenesis/issues/157#issuecomment-1250896092
	if err != nil {
		return s, err
	}
	s.RadiatorReturn, err = controller.Scale100itof(ts.client.ReadInputRegister(27)) // input reg 27 System return line temperature. visar 0 hos per
	if err != nil {
		return s, err
	}

	s.HeatCarrierForward, err = controller.Scale100itof(ts.client.ReadInputRegister(9)) // input reg 9 Condenser out temperature
	if err != nil {
		return s, err
	}
	ts.heatCarrierForward = *s.HeatCarrierForward
	s.HeatCarrierReturn, err = controller.Scale100itof(ts.client.ReadInputRegister(8)) // input reg 8 Condenser in
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

	s.HeatingAllowed = boolPointer(ts.heatingAllowed)
	s.HotwaterAllowed = boolPointer(ts.hotwaterAllowed)

	// input reg 147 Desired temperature distribution circuit Mix valve 1 verkar vara nuvarande uträknade börvärde? tex 38.08
	// input reg 1 Currently running: First prioritised demand *1
	//  1: Manual operation, 2: Defrost, 3: Hot water, 4: Heat, 5: Cool, 6: Pool, 7: Anti legionella, 98: Standby 99: No demand 100: OFF
	// input reg 18 System supply line calculated set point just nu 47.18
	// write single coil(5) enable heat 9
	// write single coil(5) enable tap water 8
	// inputreg 4 avilable gears == 12
	// inputreg 61 curren gear == 5

	// operational mode WriteSingleRegister 0 1: OFF, 2: Standby, 3: ON/Auto

	return s, nil
}

func (ts *Thermiagenesis) allowHeatpump(price float64) bool {
	if ts.heatCarrierForward == 0.0 { // vi har inte hämtat aktuell framledning ännu.
		return true
	}

	// 3.45+0.098*(60-<curtemp>)
	// calculated from 3.45 at 60C and 5.9 at 35C
	ts.calculatedCOP = 3.45 + 0.098*(60.0-ts.heatCarrierForward)

	allow := price/ts.calculatedCOP < ts.cloudConfig.DistrictHeatingPrice
	logrus.WithFields(logrus.Fields{
		"cop":           ts.calculatedCOP,
		"price":         price,
		"districtprice": ts.cloudConfig.DistrictHeatingPrice,
		"allow":         allow,
	}).Infof("thermiagenesis: allowHeatpump")
	return allow
}

func (ts *Thermiagenesis) Reconcile(current *config.HourConfig) error {

	if ts.cloudConfig.DistrictHeatingPrice == 0.0 { // control based on levels.
		ts.heatingAllowed = current.Heating
		ts.hotwaterAllowed = current.Hotwater
		logrus.WithFields(logrus.Fields{"heating": current.Heating, "hotwater": current.Hotwater}).Debugf("thermiagenesis: Reconcile")
		err := ts.allowHeating(current.Heating)
		if err != nil {
			return err
		}

		err = ts.allowHotwater(current.Hotwater)
		if err != nil {
			return err
		}
	} else { // control based on DistrictHeatingPrice
		allow := ts.allowHeatpump(current.Price)
		ts.heatingAllowed = allow
		ts.hotwaterAllowed = allow
		err := ts.allowHeating(allow)
		if err != nil {
			return err
		}
		err = ts.allowHotwater(allow)
		if err != nil {
			return err
		}
	}

	return ts.boostHotwater(current.HotwaterForce)
}

func (ts *Thermiagenesis) allowHeating(b bool) error {
	_, err := ts.client.WriteSingleCoil(9, modbusclient.CoilValue(b))
	return err
}

func (ts *Thermiagenesis) allowHotwater(b bool) error {
	_, err := ts.client.WriteSingleCoil(8, modbusclient.CoilValue(b))
	return err
}

func (ts *Thermiagenesis) boostHotwater(b bool) error {
	start := ts.cloudConfig.HotWaterNormalStartTemperature
	stop := ts.cloudConfig.HotWaterNormalStopTemperature
	if b {
		start = ts.cloudConfig.HotWaterBoostStartTemperature
		stop = ts.cloudConfig.HotWaterBoostStopTemperature
	}

	if stop == 0 || start == 0 {
		return fmt.Errorf("start/stop temperature for boost not configured")
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

func (ts *Thermiagenesis) GetHeatCurve() ([]float64, error) {
	// 5 Comfort wheel setting
	// 6 Set point heat curve, Y-coordinate 1 (highest outdoor temperature)
	// 7 Set point heat curve, Y-coordinate 2
	// 8 Set point heat curve, Y-coordinate 3
	// 9 Set point heat curve, Y-coordinate 4
	// 10 Set point heat curve, Y-coordinate 5
	// 11 Set point heat curve, Y-coordinate 6
	// 12 Set point heat curve, Y-coordinate 7 (lowest outdoor temperature)

	data, err := ts.client.ReadHoldingRegisterRaw(6, 7)
	if err != nil {
		return nil, err
	}

	return decodeHeatCurve(data), nil
}

func decodeHeatCurve(data []byte) []float64 {
	return []float64{
		float64(modbusclient.Decode(data[0:2])) / 100.0,
		float64(modbusclient.Decode(data[2:4])) / 100.0,
		float64(modbusclient.Decode(data[4:6])) / 100.0,
		float64(modbusclient.Decode(data[6:8])) / 100.0,
		float64(modbusclient.Decode(data[8:10])) / 100.0,
		float64(modbusclient.Decode(data[10:12])) / 100.0,
		float64(modbusclient.Decode(data[12:14])) / 100.0,
	}
}

/*
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 5
raw response: 0x07 0x6c (length: 2)
2025/01/27 20:36:40 value is:  19
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 6
raw response: 0x07 0x6c (length: 2)
2025/01/27 20:36:49 value is:  19
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 7
raw response: 0x0a 0x28 (length: 2)
2025/01/27 20:36:51 value is:  26
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 8
raw response: 0x0c 0x1c (length: 2)
2025/01/27 20:37:28 value is:  31
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 9
raw response: 0x0d 0xac (length: 2)
2025/01/27 20:37:29 value is:  35
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 10
raw response: 0x0e 0xd8 (length: 2)
2025/01/27 20:37:33 value is:  38
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 11
raw response: 0x11 0x94 (length: 2)
2025/01/27 20:37:35 value is:  45
root@raspberrypi4-64:/home# ./modc -addr 172.16.61.47:502 -holdingreg 12
raw response: 0x14 0x50 (length: 2)
2025/01/27 20:37:37 value is:  52
*/

func (ts *Thermiagenesis) ReconcileFromMeter(data meter.Data) error {
	// TODO
	return nil
}

func boolPointer(v bool) *bool {
	return &v
}
