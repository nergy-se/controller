package controller

import (
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

	// list active alarms
	Alarms() ([]string, error)
}

func Scale100itof(i int, err error) (*float64, error) {
	f := float64(i) / 100.0
	return &f, err
}
func Scale10itof(i int, err error) (*float64, error) {
	f := float64(i) / 10.0
	return &f, err
}

func Scale1itof(i int, err error) (*float64, error) {
	f := float64(i)
	return &f, err
}
