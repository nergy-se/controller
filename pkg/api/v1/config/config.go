package config

import (
	"sort"
	"sync"
	"time"
)

type HourConfig struct {
	Time          time.Time `json:"time"`
	Price         float64   `json:"price"`
	Hotwater      bool      `json:"hotwater"`
	HotwaterForce bool      `json:"hotwaterForce"`
	Heating       bool      `json:"heating"`
}
type Schedule map[time.Time]*HourConfig

type Config struct {
	schedule Schedule
	mutex    sync.Mutex
}

func NewConfig() *Config {
	return &Config{
		schedule: make(map[time.Time]*HourConfig),
	}
}

func (s *Config) MergeSchedule(schedule Schedule) {
	s.mutex.Lock()
	for _, hour := range schedule {
		s.schedule[hour.Time] = hour
	}
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
	return s.current(time.Now())
}
func (s *Config) current(when time.Time) *HourConfig {
	priceRange := 60 * time.Minute
	if s.isQuarterPrices() {
		priceRange = 15 * time.Minute
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for t, hour := range s.schedule {
		if inTimeSpan(t, t.Add(priceRange), when) {
			return hour
		}
	}

	return nil
}

func (s *Config) isQuarterPrices() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for t := range s.schedule {
		if t.Minute() == 15 {
			return true
		}
	}

	return false
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end) && !check.Equal(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
