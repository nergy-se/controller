package controller

import "github.com/nergy-se/controller/pkg/state"

type Controller interface {
	BlockHeating() error
	AllowHeating() error

	BlockHotwater() error
	AllowHotwater() error

	BoostHotwater() error

	//TODO do we need this?
	// SetTemp(temp float64) error

	// fetch state. Used for metrics to cloud
	State() (*state.State, error)
}
