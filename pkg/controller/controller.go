package controller

import (
	"bytes"
	"encoding/binary"

	"github.com/nergy-se/controller/pkg/state"
)

type Controller interface {
	AllowHeating(bool) error
	AllowHotwater(bool) error
	BoostHotwater(bool) error

	//TODO do we need this?
	// SetTemp(temp float64) error

	// fetch state. Used for metrics to cloud
	State() (*state.State, error)
}

func Scale100itof(i int, err error) (*float64, error) {
	f := float64(i) / 100.0
	return &f, err
}
func Scale10itof(i int, err error) (*float64, error) {
	f := float64(i) / 10.0
	return &f, err
}

func Decode(data []byte) int {
	var i int16
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &i)
	return int(i)
}
