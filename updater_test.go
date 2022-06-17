package selfupdate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_DelayUntilNextTriggerAt(t *testing.T) {
	now := time.Now()

	hourly := delayUntilNextTriggerAt(Hourly, time.Date(0, 0, 0, 0, 42, 7, 9990000, time.Local))
	hourlyTime := now.Add(hourly)
	maxHour := now.Add(2 * time.Hour)
	assert.Greater(t, hourlyTime.UnixNano(), now.UnixNano())
	assert.Less(t, hourlyTime.UnixNano(), maxHour.UnixNano())

	daily := delayUntilNextTriggerAt(Daily, time.Date(0, 0, 0, 3, 42, 7, 9990000, time.Local))
	dailyTime := now.Add(daily)
	maxDay := now.Add(48 * time.Hour)
	assert.Greater(t, dailyTime.UnixNano(), now.UnixNano())
	assert.Less(t, dailyTime.UnixNano(), maxDay.UnixNano())

	monthly := delayUntilNextTriggerAt(Monthly, time.Date(0, 0, 0, 0, 42, 7, 9990000, time.Local))
	t.Log(monthly)
	monthlyTime := now.Add(monthly)
	maxMonth := now.Add(2 * 31 * 24 * time.Hour)
	assert.Greater(t, monthlyTime.UnixNano(), now.UnixNano())
	assert.Less(t, monthlyTime.UnixNano(), maxMonth.UnixNano())
}

func Test_DelayUntilNextTriggerAtDifferentLocatrion(t *testing.T) {
	now := time.Now()

	hourly := delayUntilNextTriggerAt(Hourly, time.Date(0, 0, 0, 0, 42, 7, 9990000, time.UTC))
	t.Log(hourly)
	hourlyTime := now.Add(hourly)
	maxHour := now.Add(2 * time.Hour)
	assert.Greater(t, hourlyTime.UnixNano(), now.UnixNano())
	assert.Less(t, hourlyTime.UnixNano(), maxHour.UnixNano())
}
