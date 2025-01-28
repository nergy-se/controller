package dummy

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/api/v1/meter"
	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

type Dummy struct {
	alarms []string
	sync.Mutex
}

func New(ctx context.Context) *Dummy {
	dummy := &Dummy{}

	mux := http.NewServeMux()
	mux.HandleFunc("/alarm", func(w http.ResponseWriter, req *http.Request) {
		msg := req.URL.Query().Get("message")
		if msg == "" {
			dummy.Lock()
			fmt.Fprintf(w, "active alarms: %s", strings.Join(dummy.alarms, "|"))
			dummy.Unlock()
			return
		}
		logrus.Infof("adding alarm with %s", msg)
		dummy.Lock()
		dummy.alarms = append(dummy.alarms, msg)
		dummy.Unlock()
		fmt.Fprintf(w, "adding alarm with %s\n", msg)
	})
	mux.HandleFunc("/resetalarms", func(w http.ResponseWriter, req *http.Request) {
		dummy.Lock()
		dummy.alarms = nil
		dummy.Unlock()
		fmt.Fprintf(w, "alarms reset\n")
	})
	srv := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Error(err)
		}
	}()

	go func() {
		<-ctx.Done()
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); !errors.Is(err, http.ErrServerClosed) && err != nil {
			logrus.Error(err)
		}
	}()

	return dummy
}
func Pointer[K any](val K) *K {
	return &val
}

func (ts *Dummy) State() (*state.State, error) {
	compressor := float64(rand.Intn(100-20) + 20)
	s := &state.State{
		Indoor:             Pointer(21.1),
		Outdoor:            Pointer(11.1),
		HeatCarrierForward: nil,
		HeatCarrierReturn:  nil,
		RadiatorForward:    nil,
		RadiatorReturn:     nil,
		BrineIn:            Pointer(4.2),
		BrineOut:           Pointer(2.2),
		HotGasCompressor:   nil,
		WarmWater:          nil,
		Compressor:         &compressor,
		Alarm:              nil,
		SwitchValve:        nil,
		PumpBrine:          nil,
		PumpHeat:           nil,
		PumpRadiator:       nil,
		HeatingAllowed:     Pointer(false),
		HotwaterAllowed:    Pointer(true),
	}

	return s, nil
}

func (ts *Dummy) Reconcile(current *config.HourConfig) error {
	err := ts.allowHeating(current.Heating)
	if err != nil {
		return err
	}

	err = ts.allowHotwater(current.Hotwater)
	if err != nil {
		return err
	}

	return ts.boostHotwater(current.HotwaterForce)
}
func (ts *Dummy) allowHeating(b bool) error {
	logrus.Info("dummy: AllowHeating: ", b)
	return nil
}

func (ts *Dummy) allowHotwater(b bool) error {
	logrus.Info("dummy: AllowHotwater: ", b)
	return nil
}

func (ts *Dummy) boostHotwater(b bool) error {
	logrus.Info("dummy: BoostHotwater: ", b)
	return nil
}

func (ts *Dummy) Alarms() ([]string, error) {
	logrus.Info("dummy: Alarms")
	ts.Lock()
	defer ts.Unlock()
	return ts.alarms, nil
}
func (ts *Dummy) ReconcileFromMeter(data meter.Data) error {
	return nil
}
func (ts *Dummy) GetHeatCurve() ([]float64, error) {
	// TODO
	logrus.Info("dummy: GetHeatCurve returning 21, 22, 23, 24, 25, 26, 27")
	return []float64{21, 22, 23, 24, 25, 26, 27}, nil
}

func (ts *Dummy) SetHeatCurve(curve []float64) error {
	var address uint16 = 6
	for _, temp := range curve {
		logrus.Infof("dummy: set address %d temp %d", address, uint16(temp*100))
		address++
	}
	return nil
}
