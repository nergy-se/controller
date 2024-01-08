package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/modbus"
	v1config "github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/api/v1/types"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/controller/dummy"
	"github.com/nergy-se/controller/pkg/controller/hogforsgst"
	"github.com/nergy-se/controller/pkg/controller/thermiagenesis"
	"github.com/nergy-se/controller/pkg/modbusclient"
	"github.com/sirupsen/logrus"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

type activeAlarms struct {
	activeAlarms []string
	sync.RWMutex
}

// Add adds string to alarm list and returns true if it was added. returns false if it already exists.
func (a *activeAlarms) Add(alarm string) bool {
	a.Lock()
	defer a.Unlock()
	for _, activeAlarm := range a.activeAlarms {
		if activeAlarm == alarm {
			return false
		}
	}

	a.activeAlarms = append(a.activeAlarms, alarm)
	return true
}

func (a *activeAlarms) Clear() bool {
	hasActive := false
	a.Lock()
	if len(a.activeAlarms) > 0 {
		hasActive = true
		a.activeAlarms = nil
	}
	a.Unlock()
	return hasActive
}

type App struct {
	wg          *sync.WaitGroup
	schedule    *v1config.Config
	cloudConfig *v1config.CloudConfig
	cliConfig   *v1config.CliConfig

	controller controller.Controller

	activeAlarms *activeAlarms

	ctx            context.Context
	stopController context.CancelFunc
}

func New(config *v1config.CliConfig) *App {
	return &App{
		wg:           &sync.WaitGroup{},
		cliConfig:    config,
		schedule:     v1config.NewConfig(),
		activeAlarms: &activeAlarms{},
	}
}

func (a *App) Start(ctx context.Context) error {
	a.ctx = ctx
	err := a.setupConfig()
	if err != nil {
		return err
	}

	err = a.setupController(ctx)
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
	err = a.cliConfig.LoadSerial()
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
	return a.syncCloudConfig(false)
}

func (a *App) syncCloudConfig(fromXFetch bool) error {
	cloudConfig, err := a.fetchConfig(fromXFetch)
	if err != nil {
		return err
	}
	a.cloudConfig = cloudConfig
	return nil
}

func (a *App) setupController(pCtx context.Context) error {
	if a.stopController != nil {
		a.stopController()
	}
	ctx, cancel := context.WithCancel(pCtx)
	a.stopController = cancel
	switch a.cloudConfig.HeatControlType {
	case types.HeatControlTypeThermiaGenesis:
		handler := modbus.NewTCPClientHandler(a.cloudConfig.Address)
		client := modbus.NewClient(handler)
		a.controller = thermiagenesis.New(modbusclient.New(client, handler.Close), false, a.cloudConfig)
		logrus.Debug("configured controller thermiagenesis")

	case types.HeatControlTypeHogforsGST:
		handler := modbus.NewTCPClientHandler(a.cloudConfig.Address)
		handler.SlaveId = 1
		client := modbus.NewClient(handler)
		a.controller = hogforsgst.New(modbusclient.New(client, handler.Close), a.cloudConfig)
		logrus.Debug("configured controller hogforsgst")

	case types.HeatControlTypeDummy:
		a.controller = dummy.New(ctx)
		logrus.Debug("configured controller dummy")
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
	logrus.Debug("scheduling first reconcile in", delay)
	for {
		select {
		case <-metricsTicker.C:
			a.doSendMetrics()
			a.doSendAlarms()
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
func (a *App) doSendAlarms() {
	err := a.sendAlarms()
	if err != nil {
		logrus.Errorf("error sendAlarms: %s", err.Error())
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
		return fmt.Errorf("no current schedule")
	}

	return a.controller.Reconcile(current)
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

	return a.do("api/controller/metrics-v1", "POST", nil, bytes.NewBuffer(body), nil)
}

func (a *App) sendAlarms() error {
	alarms, err := a.controller.Alarms()
	if err != nil {
		return err
	}

	if len(alarms) == 0 {
		hadActive := a.activeAlarms.Clear()
		if hadActive {
			return a.do("api/controller/alarms-v1", "DELETE", nil, nil, nil)
		}
		return nil
	}
	for _, alarm := range alarms {
		newAlarm := a.activeAlarms.Add(alarm)
		if !newAlarm {
			continue
		}

		body, err := json.Marshal(alarm)
		if err != nil {
			logrus.Error(err)
			continue
		}

		err = a.do("api/controller/alarm-v1", "POST", nil, bytes.NewBuffer(body), nil)
		if err != nil {
			logrus.Error(err)
		}

	}
	return nil
}

func (a *App) refreshToken() error {
	resp := &struct{ Token string }{}
	err := a.do("api/token-v1", "POST", resp, nil, nil)
	if err != nil {
		return err
	}

	a.cliConfig.SetToken(resp.Token)
	return a.cliConfig.PersistToken()
}

func (a *App) updateSchedule() error {
	schedule := make(v1config.Schedule)
	err := a.do("api/controller/schedule-v1", "GET", &schedule, nil, nil)

	a.schedule.MergeSchedule(schedule)
	a.schedule.ClearOld()
	return err
}

func (a *App) fetchConfig(fromXFetch bool) (*v1config.CloudConfig, error) {
	response := &v1config.CloudConfig{}
	var header http.Header
	if fromXFetch {
		header = make(http.Header)
		header.Set("x-fetch", "ControllerConfig")
	}
	err := a.do("api/controller/config-v1", "GET", response, nil, header)
	return response, err
}

func (a *App) do(u string, method string, dst any, body io.Reader, header http.Header) error {
	logrus.Debugf("%s to %s", method, u)
	u = fmt.Sprintf("%s/%s", a.cliConfig.Server, u)
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s := a.cliConfig.SerialID(); s != "" {
		req.Header.Set("x-serial", s)
	}
	req.Header.Add("Authorization", a.cliConfig.APIToken)

	for key, val := range header {
		req.Header[key] = val
	}

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
		return fmt.Errorf("error %s %s StatusCode: %d, body: %s", method, u, resp.StatusCode, string(body))
	}

	if resp.Header.Get("x-fetch") != "" {
		// if server indicated we need new config...
		keys := strings.Split(resp.Header.Get("x-fetch"), ",")
		logrus.Debug("got x-fetch with keys: ", keys)
		for _, key := range keys {
			if key != "ControllerConfig" { //TODO when we need more of them
				continue
			}

			err = a.syncCloudConfig(true)
			if err != nil {
				return err
			}
			err = a.setupController(a.ctx)
			if err != nil {
				return err
			}
		}
	}

	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
