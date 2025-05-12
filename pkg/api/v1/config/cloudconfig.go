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

	AllowedMinIndoorTemp float64 `json:"allowedMinIndoorTemp"`
	AllowedMaxIndoorTemp float64 `json:"allowedMaxIndoorTemp"`

	Meters []Meter `json:"meters,omitempty"`

	HeatCurveAdjust              float64   `json:"heatCurveAdjust"`
	HeatCurveControlEnabled      bool      `json:"heatCurveControlEnabled"`
	HeatCurve                    []float64 `json:"heatCurve"`
	HeatingSeasonStopTemperature float64   `json:"heatingSeasonStopTemperature"`
}

type Meter struct {
	InterfaceType string `json:"interfaceType"`
	Model         string `json:"model"`
	PrimaryID     string `json:"primaryId"`
	Address       string `json:"address"`
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
