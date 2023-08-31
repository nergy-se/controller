package dummy

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

type Dummy struct {
	alarms []string
	sync.Mutex
}

func New() *Dummy {
	dummy := &Dummy{}
	http.HandleFunc("/alarm", func(w http.ResponseWriter, req *http.Request) {
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
	http.HandleFunc("/resetalarms", func(w http.ResponseWriter, req *http.Request) {
		dummy.Lock()
		dummy.alarms = nil
		dummy.Unlock()
		fmt.Fprintf(w, "alarms reset\n")
	})

	go func() {
		err := http.ListenAndServe(":8888", nil)
		if err != nil {
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
	}

	return s, nil
}

func (ts *Dummy) AllowHeating(b bool) error {
	logrus.Info("dummy: AllowHeating: ", b)
	return nil
}

func (ts *Dummy) AllowHotwater(b bool) error {
	logrus.Info("dummy: AllowHotwater: ", b)
	return nil
}

func (ts *Dummy) BoostHotwater(b bool) error {
	// TODO
	logrus.Info("dummy: BoostHotwater: ", b)

	return nil

}

func (ts *Dummy) Alarms() ([]string, error) {
	ts.Lock()
	defer ts.Unlock()
	return ts.alarms, nil
}
