package hogforsgst

import (
	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/modbusclient"
	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

type Hogforsgst struct {
	client modbusclient.Client
	COP    float64

	cloudConfig *config.CloudConfig

	heatingAllowed  bool
	hotwaterAllowed bool
}

func New(client modbusclient.Client, cloudConfig *config.CloudConfig) *Hogforsgst {
	return &Hogforsgst{
		client:      client,
		cloudConfig: cloudConfig,
		COP:         3.5, // default lower COP should be arround 3 on thermia.
	}
}

func (ts *Hogforsgst) State() (*state.State, error) {
	s := &state.State{}
	var err error

	s.BrineIn, err = controller.Scale10itof(ts.client.ReadHoldingRegister(551))
	if err != nil {
		return s, err
	}

	s.BrineOut, err = controller.Scale10itof(ts.client.ReadHoldingRegister(553))
	if err != nil {
		return s, err
	}

	s.HeatCarrierForward, err = controller.Scale10itof(ts.client.ReadHoldingRegister(555))
	if err != nil {
		return s, err
	}

	s.PumpBrine, err = controller.Scale1itof(ts.client.ReadHoldingRegister(563))
	if err != nil {
		return s, err
	}

	//TODO
	// värmepump EL effekt just nu 1936 1 dec
	// värmepump tillförd energi just nu 975 1 dec
	// hetgas tillförd energi kw 971 1 dec
	// ex (61.9+0.9) / 20.4kw

	s.RadiatorForward, err = controller.Scale10itof(ts.client.ReadHoldingRegister(283)) // 101TE41.2 Värme framledningstemperatur
	if err != nil {
		return s, err
	}

	s.RadiatorReturn, err = controller.Scale10itof(ts.client.ReadHoldingRegister(281)) // 101TE42 Värme returtemperatur
	if err != nil {
		return s, err
	}

	s.Outdoor, err = controller.Scale10itof(ts.client.ReadHoldingRegister(275)) // 101TE00 Utetemperatur
	if err != nil {
		return s, err
	}
	gear, err := controller.Scale10itof(ts.client.ReadHoldingRegister(565))
	if err != nil {
		return s, err
	}

	s.HeatingAllowed = boolPointer(ts.heatingAllowed)
	s.HotwaterAllowed = boolPointer(ts.hotwaterAllowed)

	speed := (float64(*gear) / 10.0) * 100 // it has 10 gears
	s.Compressor = &speed

	s.COP, err = controller.Scale10itof(ts.client.ReadHoldingRegister(408))
	if err != nil {
		return nil, err
	}
	if speed > 0.0 { // dont count COP if pump isnt running
		ts.COP = *s.COP
	}

	return s, nil
}

func (ts *Hogforsgst) allowHeatpump(price float64) bool {
	//TODO kolla om vi behöver ta avg COP över längre tid?
	allow := price/ts.COP < ts.cloudConfig.DistrictHeatingPrice
	logrus.WithFields(logrus.Fields{
		"cop":           ts.COP,
		"price":         price,
		"districtprice": ts.cloudConfig.DistrictHeatingPrice,
		"allow":         allow,
	}).Infof("hogforsgst: allowHeatpump")
	return allow
}

func (ts *Hogforsgst) Reconcile(current *config.HourConfig) error {

	if !ts.allowHeatpump(current.Price) {
		ts.heatingAllowed = false
		ts.hotwaterAllowed = false
		_, err := ts.client.WriteSingleRegister(4031-1, 1) // external control true
		if err != nil {
			return err
		}
		_, err = ts.client.WriteSingleRegister(4051-1, 20) // 20 C will turn off heatpump
		if err != nil {
			return err
		}
		return nil
	}
	ts.heatingAllowed = true
	ts.hotwaterAllowed = true
	// allow heatpump normal operations.
	_, err := ts.client.WriteSingleRegister(4031-1, 0) // external control false
	if err != nil {
		return err
	}

	return nil
}

func (ts *Hogforsgst) Alarms() ([]string, error) {
	// TODO
	return nil, nil
}
func boolPointer(v bool) *bool {
	return &v
}
