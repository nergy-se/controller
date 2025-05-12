package e2e

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/fortnoxab/gohtmock"
	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/app"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tbrandon/mbserver"
)

/*
{
  "controllerId": "88e7f9b7-7a6d-41e1-9861-0817998443c7",
  "heatControlType": "thermiagenesis",
  "address": "10.0.0.1:502",
  "consideredCheap": 0,
  "electricBasePrice": 0,
  "hotWaterHours": 2,
  "maxLevelHeating": 10,
  "maxLevelHotwater": 10,
  "hotWaterBoostStartTemperature": 52,
  "hotWaterBoostStopTemperature": 58,
  "hotWaterNormalStartTemperature": 45,
  "hotWaterNormalStopTemperature": 57,
  "levelFormula": "",
  "currency": "",
  "districtHeatingPrice": 0,
  "COP": 0,
  "meters": [
    {
      "id": "b23a770a-82a5-4d2d-9d98-b1d2cbb59cb4",
      "controllerId": "88e7f9b7-7a6d-41e1-9861-0817998443c7",
      "interfaceType": "mqtt",
      "model": "p1ib",
      "position": "building",
      "primaryId": "1"
    }
  ],
  "heatCurveAdjust": 0,
  "heatCurve": [
    19,
    26,
    31,
    35,
    38,
    45,
    52
  ]
}
*/

func TestThermiaSendCurrentConfig(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	var tests = []struct {
		name                 string
		currentPumpCurve     []float64
		expectedCurveConfig  []float64
		adjust               float64
		expectedAdjustConfig float64
	}{
		{
			name:                 "test set curve +1 offset",
			currentPumpCurve:     []float64{21, 22, 23, 24, 25, 26, 27},
			adjust:               21,
			expectedCurveConfig:  []float64{20, 21, 22, 23, 24, 25, 26},
			expectedAdjustConfig: 1,
		},
		{
			name:                 "test set curve without offset",
			currentPumpCurve:     []float64{20, 21, 22, 23, 24, 25, 26},
			adjust:               20,
			expectedCurveConfig:  []float64{20, 21, 22, 23, 24, 25, 26},
			expectedAdjustConfig: 0,
		},
		{
			name:                 "test set curve -2 offset",
			currentPumpCurve:     []float64{22, 23, 24, 25, 26, 27, 28},
			adjust:               18,
			expectedCurveConfig:  []float64{24, 25, 26, 27, 28, 29, 30},
			expectedAdjustConfig: -2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := gohtmock.New()
			config := &config.CliConfig{
				Server:     mock.URL(),
				SerialFile: "/dev/null",
				APIToken:   "mysecrettoken",
			}
			app := app.New(config)

			done := make(chan bool)
			mock.Mock("/api/controller/config-v1", `
{
  "controllerId": "88e7f9b7-7a6d-41e1-9861-081799844311",
  "heatControlType": "thermiagenesis",
  "address": "127.0.0.1:1502",
  "consideredCheap": 0,
  "electricBasePrice": 0,
  "hotWaterHours": 2,
  "maxLevelHeating": 10,
  "maxLevelHotwater": 10,
  "hotWaterBoostStartTemperature": 52,
  "hotWaterBoostStopTemperature": 58,
  "hotWaterNormalStartTemperature": 45,
  "hotWaterNormalStopTemperature": 57,
  "levelFormula": "",
  "currency": "",
  "districtHeatingPrice": 0,
  "COP": 0,
  "heatCurveControlEnabled": true,
  "heatCurveAdjust": 0,
  "heatCurve": [
    19,
    26,
    31,
    35,
    38,
    45,
    52
  ],
  "heatingSeasonStopTemperature": 13
}`)

			mock.Mock("/api/controller/schedule-v1", fmt.Sprintf(`
{
  "%[1]s": {
    "time": "%[1]s",
    "price": 0.417,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  }
}`, time.Now().Format(time.RFC3339)))
			mock.Mock("/api/controller/config-v1", "", func(r *http.Request) int {
				b, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				fmt.Println("body was", string(b))
				j, err := json.Marshal(tt.expectedCurveConfig)
				assert.NoError(t, err)
				assert.Contains(t, string(b), fmt.Sprintf(`"heatCurve":%s`, string(j)))
				assert.Contains(t, string(b), fmt.Sprintf(`"heatCurveAdjust":%s`, strconv.FormatFloat(tt.expectedAdjustConfig, 'g', -1, 64)))
				assert.Contains(t, string(b), `"heatingSeasonStopTemperature":13`)
				return 200
			}).SetMethod("POST")
			mock.Mock("/api/controller/metrics-v1", "", func(r *http.Request) int {
				b, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				// fmt.Println("body was", string(b))
				assert.Contains(t, string(b), `"outdoor":-15.5`)
				assert.Contains(t, string(b), `"indoorSetpoint":`+strconv.FormatFloat(tt.adjust, 'g', -1, 64))
				assert.Contains(t, string(b), `"heatingAllowed":true,"hotwaterAllowed":true`)
				defer close(done)
				return 200
			}).SetMethod("POST")
			serv := mbserver.NewServer()
			serv.HoldingRegisters[5] = uint16(tt.adjust * 100) // comfort wheeel
			for i, temp := range tt.currentPumpCurve {
				serv.HoldingRegisters[i+6] = uint16(temp * 100) // heatcurve
			}
			serv.HoldingRegisters[16] = uint16(13 * 100) // heatingSeasonStopTemperature 13.0

			serv.InputRegisters[13] = toUint(-15.5 * 100) // outdoor temp
			err := serv.ListenTCP("127.0.0.1:1502")
			assert.NoError(t, err)
			defer serv.Close()

			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			err = app.Start(ctx)
			assert.NoError(t, err)

			<-done

			assert.Equal(t, uint16(4500), serv.HoldingRegisters[22])
			assert.Equal(t, uint16(5700), serv.HoldingRegisters[23])
			mock.AssertCallCount(t, "POST", "/api/controller/config-v1", 1)
			mock.AssertCallCount(t, "POST", "/api/controller/metrics-v1", 1)
			mock.AssertMocksCalled(t)
		})
	}
}
func TestThermiaChangeConfigFromCloud(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	var tests = []struct {
		name                    string
		curve                   []float64
		expectedSetCurveOnPump  []float64
		adjust                  float64
		expectedSetAdjustOnPump float64
	}{
		{
			name:                    "test set curve +1 offset",
			curve:                   []float64{20, 26, 31, 35, 38, 45, 52},
			adjust:                  1,
			expectedSetCurveOnPump:  []float64{21, 27, 32, 36, 39, 46, 53},
			expectedSetAdjustOnPump: 21,
		},
		{
			name:                    "test set curve without offset",
			curve:                   []float64{20, 26, 31, 35, 38, 45, 52},
			adjust:                  0,
			expectedSetCurveOnPump:  []float64{20, 26, 31, 35, 38, 45, 52},
			expectedSetAdjustOnPump: 20,
		},
		{
			name:                    "test set curve -1 offset",
			curve:                   []float64{20, 26, 31, 35, 38, 45, 52},
			adjust:                  -1,
			expectedSetCurveOnPump:  []float64{19, 25, 30, 34, 37, 44, 51},
			expectedSetAdjustOnPump: 19,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := gohtmock.New()
			config := &config.CliConfig{
				Server:     mock.URL(),
				SerialFile: "/dev/null",
				APIToken:   "mysecrettoken",
			}
			app := app.New(config)

			done := make(chan bool)

			j, err := json.Marshal(tt.curve)
			assert.NoError(t, err)
			mock.Mock("/api/controller/config-v1", fmt.Sprintf(`
{
  "controllerId": "88e7f9b7-7a6d-41e1-9861-081799844311",
  "heatControlType": "thermiagenesis",
  "address": "127.0.0.1:1502",
  "consideredCheap": 0,
  "electricBasePrice": 0,
  "hotWaterHours": 2,
  "maxLevelHeating": 10,
  "maxLevelHotwater": 10,
  "hotWaterBoostStartTemperature": 52,
  "hotWaterBoostStopTemperature": 58,
  "hotWaterNormalStartTemperature": 45,
  "hotWaterNormalStopTemperature": 57,
  "levelFormula": "",
  "currency": "",
  "districtHeatingPrice": 0,
  "COP": 0,
  "heatCurveControlEnabled": true,
  "heatCurveAdjust": %s,
  "heatCurve": %s,
  "heatingSeasonStopTemperature": 13
}`, strconv.FormatFloat(tt.adjust, 'g', -1, 64), string(j))).Filter(func(r *http.Request) bool {
				if r.Header.Get("x-fetch") == "ControllerConfig" {
					defer close(done)
					return true
				}
				return false
			})

			mock.Mock("/api/controller/config-v1", `
{
  "controllerId": "88e7f9b7-7a6d-41e1-9861-081799844311",
  "heatControlType": "thermiagenesis",
  "address": "127.0.0.1:1502",
  "consideredCheap": 0,
  "electricBasePrice": 0,
  "hotWaterHours": 2,
  "maxLevelHeating": 10,
  "maxLevelHotwater": 10,
  "hotWaterBoostStartTemperature": 52,
  "hotWaterBoostStopTemperature": 58,
  "hotWaterNormalStartTemperature": 45,
  "hotWaterNormalStopTemperature": 57,
  "levelFormula": "",
  "currency": "",
  "districtHeatingPrice": 0,
  "COP": 0,
  "heatCurveControlEnabled": true,
  "heatCurveAdjust": 0,
  "heatCurve": [
    1,
    2,
    3,
    4,
    5,
    6,
    7
  ],
  "heatingSeasonStopTemperature": 0
}`)

			mock.Mock("/api/controller/schedule-v1", fmt.Sprintf(`
{
  "%[1]s": {
    "time": "%[1]s",
    "price": 0.417,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  }
}`, time.Now().Format(time.RFC3339)))
			mock.Mock("/api/controller/config-v1", "", func(r *http.Request) int {
				return 200
			}).SetMethod("POST")
			mock.Mock("/api/controller/metrics-v1", "", func(r *http.Request) int {
				return 200
			}).SetMethod("POST").SetHeader("x-fetch", "ControllerConfig")

			serv := mbserver.NewServer()
			err = serv.ListenTCP("127.0.0.1:1502")
			assert.NoError(t, err)
			defer serv.Close()

			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			err = app.Start(ctx)
			assert.NoError(t, err)

			<-done
			WaitFor(t, time.Second, "wait for last address 12 write", func() bool {
				return uint16(0) != serv.HoldingRegisters[12]
			})

			assert.Equal(t, int(tt.expectedSetAdjustOnPump*100), int(serv.HoldingRegisters[5]))
			for i, temp := range tt.expectedSetCurveOnPump {
				assert.Equal(t, int(temp*100), int(serv.HoldingRegisters[i+6]))
			}
			assert.Equal(t, int(13*100), int(serv.HoldingRegisters[16]))

			mock.AssertMocksCalled(t)
		})
	}
}

func toUint(i int16) uint16 {
	var arr [2]byte
	binary.BigEndian.PutUint16(arr[0:2], uint16(i))
	var result uint16
	for i := 0; i < 2; i++ {
		result = result << 8
		result += uint16(arr[i])
	}
	return result
}
func WaitFor(t *testing.T, timeout time.Duration, msg string, ok func() bool) {
	end := time.Now().Add(timeout)
	for {
		if end.Before(time.Now()) {
			t.Errorf("timeout waiting for: %s", msg)
			return
		}
		time.Sleep(10 * time.Millisecond)
		if ok() {
			return
		}
	}
}

func TestThermiaAllowedMinValues(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	mock := gohtmock.New()
	config := &config.CliConfig{
		Server:     mock.URL(),
		SerialFile: "/dev/null",
		APIToken:   "mysecrettoken",
	}
	app := app.New(config)

	done := make(chan bool)
	mock.Mock("/api/controller/config-v1", `
{
  "controllerId": "88e7f9b7-7a6d-41e1-9861-081799844311",
  "heatControlType": "thermiagenesis",
  "address": "127.0.0.1:1502",
  "consideredCheap": 0,
  "electricBasePrice": 0,
  "hotWaterHours": 2,
  "maxLevelHeating": 10,
  "maxLevelHotwater": 10,
  "hotWaterBoostStartTemperature": 52,
  "hotWaterBoostStopTemperature": 58,
  "hotWaterNormalStartTemperature": 45,
  "hotWaterNormalStopTemperature": 57,
  "levelFormula": "",
  "currency": "",
  "districtHeatingPrice": 0,
  "COP": 0,
  "heatCurveControlEnabled": true,
  "heatCurveAdjust": 0,
  "heatCurve": [
    19,
    26,
    31,
    35,
    38,
    45,
    52
  ],
  "heatingSeasonStopTemperature": 13,
  "allowedMinIndoorTemp":16,
  "allowedMinHotWaterTemp":40
}`)

	mock.Mock("/api/controller/schedule-v1", fmt.Sprintf(`
{
  "%[1]s": {
    "time": "%[1]s",
    "price": 0.417,
    "hotwater": false,
    "hotwaterForce": false,
    "heating": false
  }
}`, time.Now().Format(time.RFC3339)))
	mock.Mock("/api/controller/config-v1", "", func(r *http.Request) int {
		return 200
	}).SetMethod("POST")
	mock.Mock("/api/controller/metrics-v1", "",
		func(r *http.Request) int { // first call
			b, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(b), `"heatingAllowed":false,"hotwaterAllowed":false`)
			defer close(done)
			return 200
		}, func(r *http.Request) int { // second call
			b, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(b), `"heatingAllowed":true,"hotwaterAllowed":true`)
			defer close(done)
			return 200
		}).SetMethod("POST")

	serv := mbserver.NewServer()

	serv.InputRegisters[131] = toUint(15.0 * 100) // Indoor temp
	serv.InputRegisters[17] = toUint(39.0 * 100)  // hotwater temp
	err := serv.ListenTCP("127.0.0.1:1502")
	assert.NoError(t, err)
	defer serv.Close()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	err = app.Start(ctx)
	assert.NoError(t, err)

	<-done

	app.DoReconcile() // after second reconcile we should be guarded by allowedMinIndoorTemp and allowedMinHotWaterTemp

	assert.Equal(t, uint16(4500), serv.HoldingRegisters[22])
	assert.Equal(t, uint16(5700), serv.HoldingRegisters[23])
	mock.AssertCallCount(t, "POST", "/api/controller/config-v1", 1)
	mock.AssertCallCount(t, "POST", "/api/controller/metrics-v1", 1)
	mock.AssertMocksCalled(t)
}
