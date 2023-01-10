package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/sirupsen/logrus"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

type App struct {
	wg       *sync.WaitGroup
	schedule config.Config
	config   *config.CliConfig
}

func New(config *config.CliConfig) *App {
	return &App{
		wg:     &sync.WaitGroup{},
		config: config,
	}
}

func (a *App) Start(ctx context.Context) error {
	cloudConfig, err := a.fetchConfig()
	if err != nil {
		return err
	}
	a.schedule.SetHeatControlType(cloudConfig.HeatControlType)

	a.wg.Add(1)
	go a.controllerLoop(ctx)
	return nil
}

func (a *App) Wait() {
	a.wg.Wait()
}

func (a *App) controllerLoop(ctx context.Context) {
	defer a.wg.Done()
	delay := nextDelay()
	timer := time.NewTimer(delay)
	a.fetchSchedule()
	logrus.Debug("scheduling first run in", delay)
	for {
		select {
		case <-timer.C:
			timer.Reset(nextDelay())
			a.fetchSchedule()
		case <-ctx.Done():
			return
		}
	}
}

func (a *App) fetchSchedule() {

}

func (a *App) fetchConfig() (*config.CloudConfig, error) {
	u := fmt.Sprintf("%s/api/controller/config-v1", a.config.Server)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", a.config.APIToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching controller config StatusCode: %d", resp.StatusCode)
	}

	response := &config.CloudConfig{}
	err = json.NewDecoder(resp.Body).Decode(response)
	return response, err
}
