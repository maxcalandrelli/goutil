//
// +build windows
//
package gu_log

import (
	"fmt"

	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var (
	DEFAULT_MSGID = uint32(1)
)

type win_logger_wrapper struct {
	logger debug.Log
	level  LogLevel
	msgid  uint32
}

func (l *win_logger_wrapper) SetLevel(level LogLevel) {
	if level >= LOG_ERROR {
		l.level = level
	}
}

func (l win_logger_wrapper) GetCurrentLevel() LogLevel {
	return l.level
}

func (l win_logger_wrapper) MessageId(msgid uint32) Logger {
	return &win_logger_wrapper{logger: l.logger, level: l.level, msgid: msgid}
}

func (l win_logger_wrapper) Custom(level LogLevel, msg string, args ...interface{}) {
	if l.level >= level {
		m := fmt.Sprintf(msg, args...)
		switch level {
		case LOG_ERROR:
			l.logger.Error(l.msgid, m)
		case LOG_WARNING:
			l.logger.Warning(l.msgid, m)
		case LOG_INFO:
			l.logger.Info(l.msgid, m)
		default:
			l.logger.Info(l.msgid, fmt.Sprintf("[%s] %s", level.LogString(), m))
		}
	}
}

func (l win_logger_wrapper) Debug(msg string, args ...interface{}) {
	l.Custom(LOG_DEBUG, msg)
}

func (l win_logger_wrapper) Info(msg string, args ...interface{}) {
	if l.level >= LOG_INFO {
		l.logger.Info(l.msgid, fmt.Sprintf(msg, args...))
	}
}

func (l win_logger_wrapper) Warning(msg string, args ...interface{}) {
	if l.level >= LOG_WARNING {
		l.logger.Warning(l.msgid, fmt.Sprintf(msg, args...))
	}
}

func (l win_logger_wrapper) Error(msg string, args ...interface{}) {
	if l.level >= LOG_ERROR {
		l.logger.Error(l.msgid, fmt.Sprintf(msg, args...))
	}
}

func GetWindowsEventLogger(logger debug.Log) Logger {
	return &win_logger_wrapper{logger: logger, level: LOG_INFO, msgid: DEFAULT_MSGID}
}

func InstallWindowsLogSource(name string) error {
	return eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
}

func GetWindowsEventLoggerForSource(name string) (Logger, error) {
	elog, err := eventlog.Open(name)
	if err != nil {
		return nil, err
	}
	return GetWindowsEventLogger(elog), nil
}
