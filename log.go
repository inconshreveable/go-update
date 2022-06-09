package selfupdate

// LogError will be called to log any reason that have prevented an executable update
var LogError func(string, ...interface{}) = nil

// LogInfo will be called to log any reason that prevented an executable update due to a "user" decision via one of the callback
var LogInfo func(string, ...interface{}) = nil

// LogDebug will be called to log any reason that prevented an executable update, because there wasn't any available detected
var LogDebug func(string, ...interface{}) = nil

func logError(format string, p ...interface{}) {
	if LogError == nil {
		return
	}
	LogError(format, p...)
}

func logInfo(format string, p ...interface{}) {
	if LogInfo == nil {
		return
	}
	LogInfo(format, p...)
}

func logDebug(format string, p ...interface{}) {
	if LogDebug == nil {
		return
	}
	LogDebug(format, p...)
}
