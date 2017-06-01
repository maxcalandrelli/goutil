package gu_log

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	LOG_ERROR = iota
	LOG_WARNING
	LOG_INFO
	LOG_DEBUG
)

type LogLevel uint

var (
	NIL_LOGGER Logger = &nil_logger{}
	STD_LOGGER Logger = StdLogger()

	DebugLog Logger = NIL_LOGGER
)

func (l *LogLevel) Set(value string) error {
	s := strings.ToLower(value)
	switch s {
	case "error":
		*l = LOG_ERROR
	case "warning":
		*l = LOG_WARNING
	case "info":
		*l = LOG_INFO
	case "debug":
		*l = LOG_DEBUG
	default:
		if strings.HasPrefix(s, "debug") {
			s = strings.Replace(s, "debug", "", 1)
		}
		detail, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}
		*l = LogLevel(detail + LOG_DEBUG)
	}
	return nil
}

func (l LogLevel) String() string {
	switch l {
	case LOG_ERROR:
		return "ERROR"
	case LOG_WARNING:
		return "WARNING"
	case LOG_INFO:
		return "INFO"
	case LOG_DEBUG:
		return "DEBUG"
	}
	return fmt.Sprintf("DBG%d", l-LOG_DEBUG)
}

type Logger interface {
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	Debug(msg string)
	Custom(level LogLevel, msg string)
	SetLevel(level LogLevel)
	GetCurrentLevel() LogLevel
}

type std_logger_wrapper struct {
	logger *log.Logger
	level  LogLevel
}

func (l *std_logger_wrapper) SetLevel(level LogLevel) {
	if level >= LOG_ERROR {
		l.level = level
	}
}

func (l *std_logger_wrapper) GetCurrentLevel() LogLevel {
	return l.level
}

func (l *std_logger_wrapper) Info(msg string) {
	l.Custom(LOG_INFO, msg)
}

func (l *std_logger_wrapper) Warning(msg string) {
	l.Custom(LOG_WARNING, msg)
}

func (l *std_logger_wrapper) Error(msg string) {
	l.Custom(LOG_ERROR, msg)
}

func (l *std_logger_wrapper) Debug(msg string) {
	l.Custom(LOG_DEBUG, msg)
}

func (l *std_logger_wrapper) Custom(level LogLevel, msg string) {
	if level <= l.level {
		l.logger.Println(level.String(), " ", msg)
	}
}

func StdLogger() Logger {
	return &std_logger_wrapper{logger: log.New(os.Stderr, "", log.LstdFlags), level: LOG_INFO}
}

func CustomLogger(logger *log.Logger) Logger {
	return &std_logger_wrapper{logger: logger, level: LOG_INFO}
}

type nil_logger struct {
	std_logger_wrapper
}

func (l *nil_logger) Custom(level LogLevel, msg string) {}
