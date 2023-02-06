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

	ts.readInputRegister(10)  // 10 brine in
	ts.readInputRegister(11)  // 11 brine in
	ts.readInputRegister(13)  // 13 Outdoor temp
	ts.readInputRegister(17)  // 17 Tap water weighted temperature
	ts.readInputRegister(54)  // Compressor speed percent scale 100
	ts.readInputRegister(121) // Room temperature sensor scale 10
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
func (ts *Thermiagenesis) BlockHeating() error {
	return nil
}
func (ts *Thermiagenesis) AllowHeating() error {
	return nil
}

func (ts *Thermiagenesis) BlockHotwater() error {
	return nil
}
func (ts *Thermiagenesis) AllowHotwater() error {
	return nil
}

func (ts *Thermiagenesis) BoostHotwater() error {
	_, err := client.WriteSingleRegister(22, uint16(start*100)) // 100 = 1c
	if err != nil {
		return fmt.Errorf("error writeTemps 22: %w", err)
	}

	_, err = client.WriteSingleRegister(23, uint16(stop*100))
	if err != nil {
		return fmt.Errorf("error writeTemps 23: %w", err)
	}
	return nil

	return nil
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
