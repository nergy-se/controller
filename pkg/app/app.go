package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/nergy-se/controller/pkg/mbus"
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

type postRequest struct {
	url  string
	body []byte
}

type App struct {
	wg          *sync.WaitGroup
	schedule    *v1config.Config
	cloudConfig *v1config.CloudConfig
	cliConfig   *v1config.CliConfig

	controller controller.Controller
	mbusClient *mbus.Mbus

	activeAlarms *activeAlarms

	sendQueue chan *postRequest

	ctx            context.Context
	stopController context.CancelFunc
}

func New(config *v1config.CliConfig) *App {
	return &App{
		wg:           &sync.WaitGroup{},
		cliConfig:    config,
		schedule:     v1config.NewConfig(),
		activeAlarms: &activeAlarms{},
		mbusClient:   mbus.New(),
		sendQueue:    make(chan *postRequest, 20000),
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

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-a.sendQueue:
				if ctx.Err() != nil {
					return
				}
				code, err := a.do(req.url, http.MethodPost, nil, bytes.NewBuffer(req.body), nil, true) // dont check x-fetch on retries.

				if err != nil {
					logrus.Errorf("error POST retry %s: %s", req.url, err)
				}
				if code != 200 {
					logrus.Warnf("error retrying %s: %d adding to retry queue again", req.url, code)
					select {
					case a.sendQueue <- &postRequest{
						url:  req.url,
						body: req.body,
					}:
					default:
						logrus.Error(fmt.Errorf("%w %w", err, ErrQueueFull))
					}
					time.Sleep(time.Second * 10)
				}
			}
		}
	}()
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
	a.sendMeterValues()
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

// reconcile makes sure heatpump are in desired state
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
	state.Time = time.Now()

	body, err := json.Marshal(state)
	if err != nil {
		return err
	}

	err = a.postWithRetry("api/controller/metrics-v1", body)
	return err
}

var ErrQueueFull = errors.New("queue full")

func (a *App) sendMeterValues() {

	if con, ok := a.controller.(*hogforsgst.Hogforsgst); ok {
		datas, err := con.MeterData()
		if err != nil {
			logrus.Errorf("error fetching hogforsgst meterdata: %s", err)
		}

		for _, data := range datas {
			body, err := json.Marshal(data)
			if err != nil {
				logrus.Errorf("error marshal %s meter %s: %s", data.Model, data.Id, err)
				continue
			}

			err = a.postWithRetry("api/controller/meter-v1", body)
			if err != nil {
				logrus.Errorf("error POST %s meter %s: %s", data.Model, data.Id, err)
				continue
			}
		}

	}

	for _, meter := range a.cloudConfig.Meters {
		if meter.InterfaceType != "mbus" {
			continue
		}
		data, err := a.mbusClient.ReadValues(meter.Model, meter.PrimaryID)
		if err != nil {
			logrus.Errorf("error fetching mbus meter %s: %s", meter.PrimaryID, err)
			continue
		}
		body, err := json.Marshal(data)
		if err != nil {
			logrus.Errorf("error marshal mbus meter %s: %s", meter.PrimaryID, err)
			continue
		}

		err = a.postWithRetry("api/controller/meter-v1", body)
		if err != nil {
			logrus.Errorf("error POST mbus meter %s: %s", meter.PrimaryID, err)
			continue
		}
	}
}

func (a *App) sendAlarms() error {
	alarms, err := a.controller.Alarms()
	if err != nil {
		return err
	}

	if len(alarms) == 0 {
		hadActive := a.activeAlarms.Clear()
		if hadActive {
			_, err := a.do("api/controller/alarms-v1", "DELETE", nil, nil, nil, false)
			return err
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

		_, err = a.do("api/controller/alarm-v1", "POST", nil, bytes.NewBuffer(body), nil, false)
		if err != nil {
			logrus.Error(err)
		}

	}
	return nil
}

func (a *App) refreshToken() error {
	resp := &struct{ Token string }{}
	_, err := a.do("api/token-v1", "POST", resp, nil, nil, false)
	if err != nil {
		return err
	}

	a.cliConfig.SetToken(resp.Token)
	return a.cliConfig.PersistToken()
}

func (a *App) updateSchedule() error {
	schedule := make(v1config.Schedule)
	_, err := a.do("api/controller/schedule-v1", "GET", &schedule, nil, nil, false)

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
	_, err := a.do("api/controller/config-v1", "GET", response, nil, header, false)
	return response, err
}

func (a *App) postWithRetry(u string, body []byte) error {
	code, err := a.do(u, http.MethodPost, nil, bytes.NewBuffer(body), nil, false)

	if code != 200 {
		logrus.Warnf("error %s: %d adding to retry queue", u, code)
		select {
		case a.sendQueue <- &postRequest{
			url:  u,
			body: body,
		}:
		default:
			return fmt.Errorf("%w %w", err, ErrQueueFull)
		}
	}

	return err
}

func (a *App) do(u string, method string, dst any, body io.Reader, header http.Header, disableXFetch bool) (int, error) {
	logrus.Debugf("%s to %s", method, u)
	u = fmt.Sprintf("%s/%s", a.cliConfig.Server, u)
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return 0, err
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
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}
		return resp.StatusCode, fmt.Errorf("error %s %s StatusCode: %d, body: %s", method, u, resp.StatusCode, string(body))
	}

	if resp.Header.Get("x-fetch") != "" && !disableXFetch {
		// if server indicated we need new config...
		keys := strings.Split(resp.Header.Get("x-fetch"), ",")
		logrus.Debug("got x-fetch with keys: ", keys)
		for _, key := range keys {
			if key != "ControllerConfig" { //TODO when we need more of them
				continue
			}

			err = a.syncCloudConfig(true)
			if err != nil {
				return resp.StatusCode, err
			}
			err = a.setupController(a.ctx)
			if err != nil {
				return resp.StatusCode, err
			}
			// TODO reconnect mbus stuff here aswell? or not needed with serial tty?
		}
	}

	if dst == nil {
		return resp.StatusCode, nil
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(dst)
}
