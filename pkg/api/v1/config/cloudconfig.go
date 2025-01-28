package config

import (
	"github.com/nergy-se/controller/pkg/api/v1/types"
)

type CloudConfig struct {
	ControllerId string `json:"controllerId"`

	HeatControlType      types.HeatControlType `json:"heatControlType"`
	Address              string                `json:"address"`
	DistrictHeatingPrice float64               `json:"districtHeatingPrice"`

	HotWaterBoostStartTemperature  int64 `json:"hotWaterBoostStartTemperature"`
	HotWaterBoostStopTemperature   int64 `json:"hotWaterBoostStopTemperature"`
	HotWaterNormalStartTemperature int64 `json:"hotWaterNormalStartTemperature"`
	HotWaterNormalStopTemperature  int64 `json:"hotWaterNormalStopTemperature"`

	Meters []Meter `json:"meters,omitempty"`

	HeatCurveAdjust float64   `json:"heatCurveAdjust"`
	HeatCurve       []float64 `json:"heatCurve"`
}

type Meter struct {
	InterfaceType string `json:"interfaceType"`
	Model         string `json:"model"`
	PrimaryID     string `json:"primaryId"`
}

func CloudConfigNeedsControllerSetup(old *CloudConfig, new *CloudConfig) bool {

	if old == nil {
		return true
	}
	if old.HeatControlType != new.HeatControlType {
		return true
	}
	if old.Address != new.Address {
		return true
	}
	return false
}
