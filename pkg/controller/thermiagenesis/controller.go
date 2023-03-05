package thermiagenesis

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/goburrow/modbus"
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

	s.BrineIn, err = scale100itof(ts.readInputRegister(10)) // 10 brine in scale 100
	if err != nil {
		return s, err
	}

	s.BrineOut, err = scale100itof(ts.readInputRegister(11)) // 11 brine out scale 100
	if err != nil {
		return s, err
	}
	s.Outdoor, err = scale100itof(ts.readInputRegister(13)) // 13 Outdoor temp scale 100
	if err != nil {
		return s, err
	}

	s.WarmWater, err = scale100itof(ts.readInputRegister(17)) // 17 tank Tap water weighted temperature scale 100
	if err != nil {
		return s, err
	}
	s.Compressor, err = scale100itof(ts.readInputRegister(54)) // Compressor speed percent scale 100
	if err != nil {
		return s, err
	}
	s.Indoor, err = scale100itof(ts.readInputRegister(121)) // Room temperature sensor scale 10
	if err != nil {
		return s, err
	}
	asdf, err := scale100itof(ts.readInputRegister(12)) // System supply line temperature scale 100 vad är detta?! visar bara 200.0 så funkar nog inte? visar bara 200.0 så funkar nog inte?
	if err != nil {
		return s, err
	}
	fmt.Println("System supply line temperature: ", asdf)

	// input reg 147 Desired temperature distribution circuit Mix valve 1 verkar vara nuvarande uträknade börvärde? tex 38.08
	// input reg 7 Discharge pipe temperature verkar vara hetgasen? just nu 81.12
	// input reg 1 Currently running: First prioritised demand *1
	//  1: Manual operation, 2: Defrost, 3: Hot water, 4: Heat, 5: Cool, 6: Pool, 7: Anti legionella, 98: Standby 99: No demand 100: OFF
	// input reg 18 System supply line calculated set point just nu 47.18

	// https://github.com/CJNE/thermiagenesis/issues/157#issuecomment-1250896092
	// input reg 12 System supply line temperature visar 200 hos per
	// input reg 27 System return line temperature. visar 0 hos per
	// input reg 8 Condenser in som visar 42.08 kan vara detta? HeatCarrierReturn!
	// input reg 9 Condenser out temperature som visar 47.26 kan vara detta? HeatCarrierForward!

	// input reg 44 Brine circulation pump speed (%) just nu 66.81 PumpBrine
	// input reg 39 Condenser circulation pump speed (%) just nu 60.1 PumpHeat

	/*
		start, err := ts.readHoldingRegister(22)
		if err != nil {
			return
		}
		start /= 100

		stop, err := ts.readHoldingRegister(23)
		if err != nil {
			return
		}
		stop /= 100
	*/
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

func scale100itof(i int, err error) (*float64, error) {
	f := float64(i) / 100.0
	return &f, err
}

func decode(data []byte) int {
	var i int16
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &i)
	return int(i)
}
func (ts *Thermiagenesis) readInputRegister(address uint16) (int, error) {
	b, err := ts.client.ReadInputRegisters(address, 1)
	return decode(b), err
}

func (ts *Thermiagenesis) readHoldingRegister(address uint16) (int, error) {
	b, err := ts.client.ReadHoldingRegisters(address, 1)
	return decode(b), err
}
