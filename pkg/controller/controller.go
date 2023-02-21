package controller

import "github.com/nergy-se/controller/pkg/state"

type Controller interface {
	AllowHeating(bool) error
	AllowHotwater(bool) error
	BoostHotwater(bool) error

	//TODO do we need this?
	// SetTemp(temp float64) error

	// fetch state. Used for metrics to cloud
	State() (*state.State, error)
}
