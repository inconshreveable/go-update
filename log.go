package selfupdate

var LogError func(string, ...interface{}) = nil
var LogInfo func(string, ...interface{}) = nil

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
