package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/goburrow/modbus"
	v1config "github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/api/v1/types"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/controller/dummy"
	"github.com/nergy-se/controller/pkg/controller/thermiagenesis"
	"github.com/sirupsen/logrus"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

type App struct {
	wg          *sync.WaitGroup
	schedule    *v1config.Config
	cloudConfig *v1config.CloudConfig
	cliConfig   *v1config.CliConfig

	controller controller.Controller
}

func New(config *v1config.CliConfig) *App {
	return &App{
		wg:        &sync.WaitGroup{},
		cliConfig: config,
		schedule:  v1config.NewConfig(),
	}
}

func (a *App) Start(ctx context.Context) error {
	err := a.setupConfig()
	if err != nil {
		return err
	}

	err = a.setupController()
	if err != nil {
		return err
	}

	a.wg.Add(1)
	go a.controllerLoop(ctx)
	return nil
}

func (a *App) setupConfig() error {

	err := a.cliConfig.LoadToken()
	if err != nil {
		return err
	}

	if a.cliConfig.ControllerType != "" && a.cliConfig.Address != "" {
		logrus.Infof("using controller %s specified in cli config", a.cliConfig.ControllerType)
		a.cloudConfig = &v1config.CloudConfig{
			HeatControlType: types.HeatControlType(a.cliConfig.ControllerType),
			Address:         a.cliConfig.Address,
		}
		return nil
	}
	cloudConfig, err := a.fetchConfig()
	if err != nil {
		return err
	}
	a.cloudConfig = cloudConfig
	return nil
}

func (a *App) setupController() error {
	switch a.cloudConfig.HeatControlType {
	case types.HeatControlTypeThermiaGenesis:
		client := modbus.TCPClient(a.cloudConfig.Address)
		a.controller = thermiagenesis.New(client)
	case types.HeatControlTypeDummy:
		a.controller = dummy.New()
	}
	return nil
}

func (a *App) Wait() {
	a.wg.Wait()
}

func (a *App) controllerLoop(ctx context.Context) {
	defer a.wg.Done()
	delay := nextDelay()
	timer := time.NewTimer(delay)
	a.doUpdateSchedule()
	a.doReconcile()
	a.doSendMetrics()

	scheduleTicker := time.NewTicker(time.Hour * 6)
	refreshToken := time.NewTicker(time.Hour * 24)
	metricsTicker := time.NewTicker(time.Second * 30)
	logrus.Debug("scheduling first run in", delay)
	for {
		select {
		case <-metricsTicker.C:
			a.doSendMetrics()
		case <-timer.C:
			timer.Reset(nextDelay())
			a.doReconcile()
		case <-scheduleTicker.C:
			a.doUpdateSchedule()
		case <-refreshToken.C:
			a.doRefreshToken()
		case <-ctx.Done():
			return
		}
	}
}

func (a *App) doRefreshToken() {
	err := a.refreshToken()
	if err != nil {
		logrus.Errorf("error refreshToken: %s", err.Error())
	}
}

func (a *App) doSendMetrics() {
	err := a.sendMetrics()
	if err != nil {
		logrus.Errorf("error sendMetrics: %s", err.Error())
	}
}

func (a *App) doReconcile() {
	err := a.reconcile()
	if err != nil {
		logrus.Errorf("error reconcile: %s", err.Error())
	}
}
func (a *App) doUpdateSchedule() {
	err := a.updateSchedule()
	if err != nil {
		logrus.Errorf("error updateSchedule: %s", err.Error())
	}
}

// makes sure heatpump are in desired state
func (a *App) reconcile() error {
	logrus.Debug("reconcile heatpump")
	current := a.schedule.Current()

	if current == nil {
		return fmt.Errorf("found no curent schedule")
	}

	err := a.controller.AllowHeating(current.Heating)
	if err != nil {
		return err
	}

	// TODO make it a config to chose if we want to block both hotwater and heating?
	err = a.controller.AllowHotwater(current.Heating)
	if err != nil {
		return err
	}

	return a.controller.BoostHotwater(current.Hotwater)
}

func (a *App) sendMetrics() error {
	state, err := a.controller.State()
	if err != nil {
		return err
	}

	body, err := json.Marshal(state)
	if err != nil {
		return err
	}

	logrus.Debug("send metrics")
	return a.do("api/controller/metrics-v1", "POST", nil, bytes.NewBuffer(body))
}

func (a *App) refreshToken() error {
	logrus.Debug("refresh token")
	resp := &struct{ Token string }{}
	err := a.do("api/token-v1", "POST", resp, nil)
	if err != nil {
		return err
	}

	a.cliConfig.SetToken(resp.Token)
	logrus.Debug("refresh token done")
	return a.cliConfig.PersistToken()
}

func (a *App) updateSchedule() error {
	logrus.Debug("fetching schedule")

	schedule := make(v1config.Schedule)
	err := a.do("api/controller/schedule-v1", "GET", &schedule, nil)
	logrus.Debugf("fetched schedule: %#v ", schedule)

	a.schedule.MergeSchedule(schedule)
	a.schedule.ClearOld()
	return err
}

func (a *App) fetchConfig() (*v1config.CloudConfig, error) {
	response := &v1config.CloudConfig{}
	err := a.do("api/controller/config-v1", "GET", response, nil)
	logrus.Debugf("fetched config: %#v ", response)
	return response, err
}

func (a *App) do(u string, method string, dst any, body io.Reader) error {
	u = fmt.Sprintf("%s/%s", a.cliConfig.Server, u)
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", a.cliConfig.APIToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error fetching %s StatusCode: %d, body: %s", u, resp.StatusCode, string(body))
	}

	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
