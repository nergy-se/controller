package config

import (
	"sort"
	"sync"
	"time"

	"github.com/nergy-se/controller/pkg/api/v1/types"
)

type CliConfig struct {
	Server   string `default:"https://nergy.se"`
	APIToken string
	LogLevel string `default:"info"`
}

type HourConfig struct {
	Time     time.Time `json:"time"`
	Hotwater bool      `json:"hotwater"`
	Heating  bool      `json:"heating"`
}

type CloudConfig struct {
	HeatControlType types.HeatControlType `json:"heatControlType"`
}

type Config struct {
	heatControlType types.HeatControlType
	schedule        map[time.Time]*HourConfig
	mutex           sync.Mutex
}

func NewConfig() *Config {
	return &Config{
		schedule: make(map[time.Time]*HourConfig),
	}
}

func (s *Config) SetHeatControlType(t types.HeatControlType) {
	s.mutex.Lock()
	s.heatControlType = t
	s.mutex.Unlock()
}

func (s *Config) Add(hour *HourConfig) {
	s.mutex.Lock()
	s.schedule[hour.Time] = hour
	s.mutex.Unlock()
}

func (s *Config) ClearOld() {
	s.mutex.Lock()
	for t := range s.schedule {
		if time.Now().Add(time.Hour * -24).After(t) { // we keep the last 24h
			delete(s.schedule, t)
		}
	}
	s.mutex.Unlock()
}

func (s *Config) Last() *HourConfig {
	ss := make([]time.Time, len(s.schedule))
	i := 0
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for t := range s.schedule {
		ss[i] = t
		i++
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].After(ss[j])
	})

	return s.schedule[ss[0]]
}

func (s *Config) Current() *HourConfig {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for t, hour := range s.schedule {
		if inTimeSpan(t, t.Add(60*time.Minute), time.Now()) {
			return hour
		}
	}

	return nil
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
