package app

import "time"

func calculateNextDelay() time.Duration {
	now := time.Now()
	// Calculate the next quarter-hour mark (0, 15, 30, 45)
	nextQuarter := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		(now.Minute()/15+1)*15, // This clever math finds the next 15-min interval
		0,
		0,
		now.Location(),
	)
	return time.Until(nextQuarter)
}
