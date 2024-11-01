package alarm

import "sync"

type ActiveAlarms struct {
	activeAlarms []string
	sync.RWMutex
}

// Add adds string to alarm list and returns true if it was added. returns false if it already exists.
func (a *ActiveAlarms) Add(alarm string) bool {
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

func (a *ActiveAlarms) Clear() bool {
	hasActive := false
	a.Lock()
	if len(a.activeAlarms) > 0 {
		hasActive = true
		a.activeAlarms = nil
	}
	a.Unlock()
	return hasActive
}
