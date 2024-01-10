package config

import (
	"github.com/nergy-se/controller/pkg/api/v1/types"
)

type CloudConfig struct {
	HeatControlType                types.HeatControlType `json:"heatControlType"`
	Address                        string                `json:"address"`
	DistrictHeatingPrice           float64               `json:"districtHeatingPrice"`
	HotWaterBoostStartTemperature  int64                 `json:"hotWaterBoostStartTemperature"`
	HotWaterBoostStopTemperature   int64                 `json:"hotWaterBoostStopTemperature"`
	HotWaterNormalStartTemperature int64                 `json:"hotWaterNormalStartTemperature"`
	HotWaterNormalStopTemperature  int64                 `json:"hotWaterNormalStopTemperature"`
	ControllerId                   string                `json:"controllerId"`
	Meters                         []Meter               `json:"meters,omitempty"`
}

type Meter struct {
	InterfaceType string `json:"interfaceType"`
	Model         string `json:"model"`
	PrimaryID     string `json:"primaryId"`
}
