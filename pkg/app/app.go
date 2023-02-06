package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	v1config "github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/api/v1/types"
	"github.com/nergy-se/controller/pkg/controller"
	"github.com/nergy-se/controller/pkg/controller/thermiagenesis"
	"github.com/sirupsen/logrus"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

type App struct {
	wg        *sync.WaitGroup
	schedule  *v1config.Config
	cliConfig *v1config.CliConfig

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
	cloudConfig, err := a.fetchConfig()
	if err != nil {
		return err
	}
	a.schedule.SetHeatControlType(cloudConfig.HeatControlType)

	//TODO set a.controller based on what HeatControlType we have!

	a.wg.Add(1)
	go a.controllerLoop(ctx)
	return nil
}

func (a *App) setupController() error {
	switch a.schedule.HeatControlType() {
	case types.HeatControlTypeThermiaGenesis:
		//TODO pump IP here? fetch from server to make if configurable in web gui?
		a.controller = &thermiagenesis.Thermiagenesis{}
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
	err := a.updateSchedule()
	if err != nil {
		logrus.Error(err)
	}

	scheduleTicker := time.NewTicker(time.Hour * 6)
	metricsTicker := time.NewTicker(time.Second * 30)
	logrus.Debug("scheduling first run in", delay)
	for {
		select {
		case <-metricsTicker.C:
			logrus.Debug("send metrics")
			err := a.readDevice()
			if err != nil {
				logrus.Error(err)
				continue
			}

		case <-timer.C:
			timer.Reset(nextDelay())
		//TODO write to heatpump if needed
		// a.updateSchedule()
		case <-scheduleTicker.C:
			err := a.updateSchedule()
			if err != nil {
				logrus.Error(err)
				continue
			}

		case <-ctx.Done():
			return
		}
	}
}

func (a *App) readDevice() error {
	state, err := a.controller.State()
	if err != nil {
		return err
	}

	fmt.Println("logging state to cloud", state)
	return nil
}

func (a *App) updateSchedule() error {
	logrus.Debug("fetching schedule")

	schedule := make(v1config.Schedule)
	err := a.doFetch("api/controller/schedule-v1", &schedule)
	logrus.Debugf("fetched schedule: %#v ", schedule)

	a.schedule.MergeSchedule(schedule)
	a.schedule.ClearOld()
	return err
}

func (a *App) fetchConfig() (*v1config.CloudConfig, error) {
	response := &v1config.CloudConfig{}
	err := a.doFetch("api/controller/config-v1", response)
	logrus.Debugf("fetched config: %#v ", response)
	return response, err
}

func (a *App) doFetch(u string, dst any) error {
	u = fmt.Sprintf("%s/%s", a.cliConfig.Server, u)
	req, err := http.NewRequest("GET", u, nil)
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

	if resp.StatusCode != 200 {
		return fmt.Errorf("error fetching controller config StatusCode: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}
