package logger

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

func (l LogLevel) Valid() bool {
	switch l {
	case LogDebug, LogInfo, LogWarn, LogError:
		return true
	}
	return false
}
