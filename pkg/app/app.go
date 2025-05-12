package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/modbus"
	mqttv2 "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/nergy-se/controller/pkg/alarm"
	v1config "github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/api/v1/meter"
	"github.com/nergy-se/controller/pkg/api/v1/types"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/controller/dummy"
	"github.com/nergy-se/controller/pkg/controller/hogforsgst"
	"github.com/nergy-se/controller/pkg/controller/thermiagenesis"
	"github.com/nergy-se/controller/pkg/mbus"
	"github.com/nergy-se/controller/pkg/modbusclient"
	"github.com/nergy-se/controller/pkg/mqtt"
	"github.com/nergy-se/controller/pkg/state"
	"github.com/sirupsen/logrus"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
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

	activeAlarms *alarm.ActiveAlarms

	sendQueue chan *postRequest

	ctx            context.Context
	stopController context.CancelFunc

	mqttServer *mqttv2.Server
	meterCache *meter.Cache
	stateCache *state.Cache
}

func New(config *v1config.CliConfig) *App {
	return &App{
		wg:           &sync.WaitGroup{},
		cliConfig:    config,
		schedule:     v1config.NewConfig(),
		activeAlarms: &alarm.ActiveAlarms{},
		mbusClient:   mbus.New(),
		sendQueue:    make(chan *postRequest, 20000),
		meterCache:   &meter.Cache{},
		stateCache:   &state.Cache{},
	}
}

func (a *App) sendCurrentSettings() {
	curve, adjust, err := a.controller.GetHeatCurve()
	if err != nil {
		logrus.Errorf("error fetching heatcurve: %s", err.Error())
	}

	heatingSeasonStopTemperature, err := a.controller.GetHeatingSeasonStopTemperature()
	if err != nil {
		logrus.Errorf("error fetching  heatingSeasonStopTemperature: %s", err.Error())
	}

	type curveData struct {
		HeatCurve                    []float64 `json:"heatCurve,omitempty"`
		HeatCurveAdjust              float64   `json:"heatCurveAdjust"`
		HeatingSeasonStopTemperature float64   `json:"heatingSeasonStopTemperature,omitempty"`
	}
	body, err := json.Marshal(curveData{
		HeatCurve:                    curve,
		HeatCurveAdjust:              adjust,
		HeatingSeasonStopTemperature: heatingSeasonStopTemperature,
	})
	if err != nil {
		logrus.Errorf("error marshal curveData: %s", err)
		return
	}
	err = a.postWithRetry("api/controller/config-v1", body)
	if err != nil {
		logrus.Errorf("error sending heatcurve-v1 to cloud: %s", err.Error())
	}
}

func (a *App) Start(ctx context.Context) error {
	a.ctx = ctx
	err := a.setupInitialConfig()
	if err != nil {
		return err
	}
	a.doUpdateSchedule()

	err = a.setupController(ctx)
	if err != nil {
		return err
	}

	a.sendCurrentSettings()

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

func (a *App) setupInitialConfig() error {

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
	cloudConfig, err := a.fetchConfig("")
	if err != nil {
		return err
	}
	a.cloudConfig = cloudConfig
	err = a.StartMQTTServer(a.ctx) //TODO we use parent context here so if we would like to react on changed mqtt config we need to use ctx to it gets restarted.
	if err != nil {
		return err
	}

	return nil
}

func (a *App) syncCloudConfig(fromXFetch string) error {
	cloudConfig, err := a.fetchConfig(fromXFetch)
	if err != nil {
		return err
	}
	needsSetupController := v1config.CloudConfigNeedsControllerSetup(a.cloudConfig, cloudConfig)
	heatCurveDiff := !slices.Equal(a.cloudConfig.HeatCurve, cloudConfig.HeatCurve) || a.cloudConfig.HeatCurveAdjust != cloudConfig.HeatCurveAdjust
	heatingSeasonStopTemperatureDiff := a.cloudConfig.HeatingSeasonStopTemperature != cloudConfig.HeatingSeasonStopTemperature

	a.cloudConfig = cloudConfig

	err = a.StartMQTTServer(a.ctx) //TODO we use parent context here so if we would like to react on changed mqtt config we need to use ctx to it gets restarted.
	if err != nil {
		return err
	}

	if needsSetupController {
		err = a.setupController(a.ctx)
		if err != nil {
			logrus.Errorf("error setupController: %s", err.Error())
		}
	}

	if cloudConfig.HeatCurveControlEnabled {
		if heatCurveDiff && a.cloudConfig.HeatCurve != nil {
			err = a.controller.SetHeatCurve(a.cloudConfig.HeatCurve, a.cloudConfig.HeatCurveAdjust)
			if err != nil {
				logrus.Errorf("error SetHeatCurve: %s", err.Error())
			}
		}
		if heatingSeasonStopTemperatureDiff {
			err = a.controller.SetHeatingSeasonStopTemperature(a.cloudConfig.HeatingSeasonStopTemperature)
			if err != nil {
				logrus.Errorf("error SetHeatingSeasonStopTemperature: %s", err.Error())
			}
		}
	}

	return nil
}

func (a *App) setupController(pCtx context.Context) error {
	if a.stopController != nil {
		a.stopController()
	}
	var ctx context.Context
	ctx, a.stopController = context.WithCancel(pCtx)

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

	// We must reconcile after controller has been setup otherwise allow values in state can mismatch
	a.doReconcile()
	return nil
}

func (a *App) Wait() {
	a.wg.Wait()
}

// TODO start mqtt server if any mqtt config. then have separate function to do the Subscriptions based on whith meter (p1ib etc...)
func (a *App) StartMQTTServer(ctx context.Context) error {
	var err error
	hasAnyMQTT := false
	for _, m := range a.cloudConfig.Meters {
		if m.InterfaceType == "mqtt" {
			hasAnyMQTT = true
			if a.mqttServer == nil {
				// TODO allow incoming TCP on 1883 on controllerimage iptables rules.
				a.mqttServer, err = mqtt.Start(ctx, a.wg)
				if err != nil {
					return err
				}
				if m.Model == "p1ib" {
					err := a.mqttServer.Subscribe("p1ib/sensor_state", 1, func(cl *mqttv2.Client, sub packets.Subscription, pk packets.Packet) {
						data := &mqtt.P1ib{}
						err := json.Unmarshal(pk.Payload, data)
						if err != nil {
							logrus.Errorf("error unmarshal p1ib payload: %s", err)
							return
						}

						meterData := data.AsMeterData(m.PrimaryID)
						meterData.Time = time.Now()
						a.meterCache.Set(meterData)
					})
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if !hasAnyMQTT && a.mqttServer != nil {
		a.mqttServer.Close()
	}
	return nil
}

func (a *App) controllerLoop(ctx context.Context) {
	defer a.wg.Done()
	delay := nextDelay()
	timer := time.NewTimer(delay)
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
		if strings.Contains(err.Error(), "error fetching state:") {
			// try to get new config in case we changed it and are in a reconnect loop.
			err = a.syncCloudConfig("")
			if err != nil {
				logrus.Errorf("error syncCloudConfig: %s", err.Error())
			}
		}
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

	// indoor temp rules here
	if t := a.stateCache.Get().Indoor; t != nil && *t < a.cloudConfig.AllowedMinIndoorTemp { // dont allow cooler than the setting indoor
		current.Heating = true
	}

	return a.controller.Reconcile(current)
}

func (a *App) sendMetrics() error {
	state, err := a.controller.State()
	if err != nil {
		return fmt.Errorf("error fetching state: %w", err)
	}
	state.Time = time.Now()

	cachedState := a.stateCache.Get()
	if cachedState.Indoor != nil && cachedState.Outdoor == nil { //TODO make this more generic with more fields etc.
		state.Indoor = cachedState.Indoor
	}
	a.stateCache.Set(state)

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

	for _, m := range a.cloudConfig.Meters {
		var data *meter.Data
		var err error
		switch m.InterfaceType {
		case "mbus":
			data, err = a.mbusClient.ReadValues(m.Model, m.PrimaryID)
			if err != nil {
				logrus.Errorf("error fetching mbus meter %s: %s", m.PrimaryID, err)
				continue
			}
		case "mqtt":
			if m.Model == "p1ib" {
				data = a.meterCache.Get()
			}
		case "modbus-tcp":
			if m.Model == "holdingreg-10scale-16bit" { // TODO add m.Position = indoor_temp

				handler := modbus.NewTCPClientHandler(m.Address)
				c := modbusclient.New(modbus.NewClient(handler), handler.Close)
				id, _ := strconv.Atoi(m.PrimaryID)
				var val int
				val, err = c.ReadHoldingRegister16(uint16(id))
				handler.Close()
				if err != nil {
					logrus.Errorf("error ReadHoldingRegister16 from address:%s id:%s", m.Address, m.PrimaryID)
				}
				t := float64(val) / 10.0
				a.stateCache.Set(&state.State{Indoor: &t})
				continue // continue for loos since we have nothing to POST to meter-v1 here (send with controller metrics)
			}
		default:
			continue
		}

		if data == nil {
			logrus.Errorf("empty data for meter %s: %s id: %s", m.InterfaceType, m.Model, m.PrimaryID)
			continue
		}

		body, err := json.Marshal(data)
		if err != nil {
			logrus.Errorf("error marshal mbus meter %s: %s", m.PrimaryID, err)
			continue
		}

		err = a.postWithRetry("api/controller/meter-v1", body)
		if err != nil {
			logrus.Errorf("error POST mbus meter %s: %s", m.PrimaryID, err)
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
			_, err := a.do("api/controller/alarms-v1", "DELETE", nil, nil, nil, true)
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

		_, err = a.do("api/controller/alarm-v1", "POST", nil, bytes.NewBuffer(body), nil, true)
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

func (a *App) fetchConfig(fromXFetch string) (*v1config.CloudConfig, error) {
	response := &v1config.CloudConfig{}
	var header http.Header
	if fromXFetch != "" {
		header = make(http.Header)
		header.Set("x-fetch", fromXFetch)
	}
	_, err := a.do("api/controller/config-v1", "GET", response, nil, header, true)
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

			err := a.syncCloudConfig(key)
			if err != nil {
				logrus.Errorf("error from syncCloudConfig: %s", err.Error())
			}
		}
	}

	if dst == nil {
		return resp.StatusCode, nil
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(dst)
}
