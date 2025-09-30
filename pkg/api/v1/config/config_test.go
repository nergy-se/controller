package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuarterlySchedule(t *testing.T) {

	d := `
{
  "2025-09-30T00:00:00+02:00": {
    "time": "2025-09-30T00:00:00+02:00",
    "price": 1,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  },
  "2025-09-30T00:15:00+02:00": {
    "time": "2025-09-30T01:00:00+02:00",
    "price": 2,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  },
  "2025-09-30T00:30:00+02:00": {
    "time": "2025-09-30T02:00:00+02:00",
    "price": 3,
    "hotwater": true,
    "hotwaterForce": true,
    "heating": true
  },
  "2025-09-30T00:45:00+02:00": {
    "time": "2025-09-30T03:00:00+02:00",
    "price": 4,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  },
  "2025-09-30T01:00:00+02:00": {
    "time": "2025-09-30T03:00:00+02:00",
    "price": 5,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  }
}`

	schedule := Schedule{}
	err := json.Unmarshal([]byte(d), &schedule)
	assert.NoError(t, err)
	conf := NewConfig()
	conf.schedule = schedule

	assert.True(t, conf.isQuarterPrices())

	ts, err := time.Parse(time.RFC3339, "2025-09-30T00:00:00+02:00")
	assert.NoError(t, err)
	cur := conf.current(ts)
	assert.Equal(t, 1.0, cur.Price)

	ts, err = time.Parse(time.RFC3339, "2025-09-30T00:05:00+02:00")
	assert.NoError(t, err)
	cur = conf.current(ts)
	assert.Equal(t, 1.0, cur.Price)

	ts, err = time.Parse(time.RFC3339, "2025-09-30T00:45:00+02:00")
	assert.NoError(t, err)
	cur = conf.current(ts)
	assert.Equal(t, 4.0, cur.Price)

	ts, err = time.Parse(time.RFC3339, "2025-09-30T00:44:59+02:00")
	assert.NoError(t, err)
	cur = conf.current(ts)
	assert.Equal(t, 3.0, cur.Price)

	ts, err = time.Parse(time.RFC3339, "2025-09-30T01:00:00+02:00")
	assert.NoError(t, err)
	cur = conf.current(ts)
	assert.Equal(t, 5.0, cur.Price)
}

func TestNoQuarterlySchedule(t *testing.T) {

	d := `
{
  "2025-09-30T00:00:00+02:00": {
    "time": "2025-09-30T00:00:00+02:00",
    "price": 1,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  },
  "2025-09-30T01:00:00+02:00": {
    "time": "2025-09-30T01:00:00+02:00",
    "price": 2,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  },
  "2025-09-30T02:00:00+02:00": {
    "time": "2025-09-30T02:00:00+02:00",
    "price": 3,
    "hotwater": true,
    "hotwaterForce": true,
    "heating": true
  },
  "2025-09-30T03:00:00+02:00": {
    "time": "2025-09-30T03:00:00+02:00",
    "price": 4,
    "hotwater": true,
    "hotwaterForce": false,
    "heating": true
  }
}`

	schedule := Schedule{}
	err := json.Unmarshal([]byte(d), &schedule)
	assert.NoError(t, err)
	conf := NewConfig()
	conf.schedule = schedule

	assert.False(t, conf.isQuarterPrices())

	ts, err := time.Parse(time.RFC3339, "2025-09-30T00:00:00+02:00")
	assert.NoError(t, err)
	cur := conf.current(ts)
	assert.Equal(t, 1.0, cur.Price)

	ts, err = time.Parse(time.RFC3339, "2025-09-30T00:59:00+02:00")
	assert.NoError(t, err)
	cur = conf.current(ts)
	assert.Equal(t, 1.0, cur.Price)

	ts, err = time.Parse(time.RFC3339, "2025-09-30T01:00:00+02:00")
	assert.NoError(t, err)
	cur = conf.current(ts)
	assert.Equal(t, 2.0, cur.Price)

}
