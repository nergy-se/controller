package hogforsgst

import (
	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/modbusclient"
	"github.com/nergy-se/controller/pkg/state"
)

type Hogforsgst struct {
	client modbusclient.Client
	COP    float64

	cloudConfig *config.CloudConfig
}

func New(client modbusclient.Client, cloudConfig *config.CloudConfig) *Hogforsgst {
	return &Hogforsgst{
		client:      client,
		cloudConfig: cloudConfig,
	}
}

func (ts *Hogforsgst) State() (*state.State, error) {
	s := &state.State{}
	var err error

	s.BrineIn, err = controller.Scale10itof(ts.client.ReadHoldingRegister(551 - 1))
	if err != nil {
		return s, err
	}

	s.BrineOut, err = controller.Scale10itof(ts.client.ReadHoldingRegister(553 - 1))
	if err != nil {
		return s, err
	}

	s.HeatCarrierForward, err = controller.Scale10itof(ts.client.ReadHoldingRegister(555 - 1))
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

	speed := (float64(*gear) / 10.0) * 100 // it has 10 gears
	s.Compressor = &speed

	s.COP, err = controller.Scale10itof(ts.client.ReadHoldingRegister(408))
	if err != nil {
		return nil, err
	}
	ts.COP = *s.COP
	return s, nil
}

func (ts *Hogforsgst) allowHeatpump() bool {
	//TODO kolla om vi behöver ta avg COP över längre tid?
	price := 1.0 //TODO
	return price/ts.COP < ts.cloudConfig.DistrictHeatingPrice
}

func (ts *Hogforsgst) AllowHeating(b bool) error {
	// _, err := ts.client.WriteSingleRegister(9, modbusclient.CoilValue(b))
	// return err
	return nil
}

func (ts *Hogforsgst) AllowHotwater(b bool) error {
	// _, err := ts.client.WriteSingleRegister(8, modbusclient.CoilValue(b))
	// return err
	return nil
}

func (ts *Hogforsgst) BoostHotwater(b bool) error {
	// TODO do we even want this here?
	return nil
}

func (ts *Hogforsgst) Alarms() ([]string, error) {
	// TODO
	return nil, nil
}
