package gu_log

const (
	LOG_ERROR = iota
	LOG_WARNING
	LOG_INFO
	LOG_DEBUG
)

var (
	NIL_LOGGER Logger = &nil_logger{}
	STD_LOGGER Logger = StdLogger()

	DebugLog Logger = NIL_LOGGER
)

type Logger interface {
	Info(msg string, args ...interface{})
	Warning(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Custom(level LogLevel, msg string, args ...interface{})
	SetLevel(level LogLevel)
	GetCurrentLevel() LogLevel
	MessageId(id uint32) Logger
}
