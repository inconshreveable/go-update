package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	debugCalled := false
	errorCalled := false
	infoCalled := false

	logError("error")
	assert.False(t, errorCalled)
	logInfo("info")
	assert.False(t, infoCalled)
	logDebug("debug")
	assert.False(t, debugCalled)

	LogDebug = func(s string, i ...interface{}) {
		if s == "debug" {
			debugCalled = true
		}
	}
	LogError = func(s string, i ...interface{}) {
		if s == "error" {
			errorCalled = true
		}
	}
	LogInfo = func(s string, i ...interface{}) {
		if s == "info" {
			infoCalled = true
		}
	}

	logError("error")
	assert.True(t, errorCalled)
	logInfo("info")
	assert.True(t, infoCalled)
	logDebug("debug")
	assert.True(t, debugCalled)
}
