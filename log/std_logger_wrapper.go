package gu_log

import (
	"fmt"
	"log"
	"os"
)

type std_logger_wrapper struct {
	logger *log.Logger
	level  LogLevel
	msgid  uint32
}

func (l *std_logger_wrapper) SetLevel(level LogLevel) {
	if level >= LOG_ERROR {
		l.level = level
	}
}

func (l *std_logger_wrapper) GetCurrentLevel() LogLevel {
	return l.level
}

func (l *std_logger_wrapper) Info(msg string, args ...interface{}) {
	l.Custom(LOG_INFO, msg, args...)
}

func (l *std_logger_wrapper) Warning(msg string, args ...interface{}) {
	l.Custom(LOG_WARNING, msg, args...)
}

func (l *std_logger_wrapper) Error(msg string, args ...interface{}) {
	l.Custom(LOG_ERROR, msg, args...)
}

func (l *std_logger_wrapper) Debug(msg string, args ...interface{}) {
	l.Custom(LOG_DEBUG, msg, args...)
}

func (l *std_logger_wrapper) Custom(level LogLevel, msg string, args ...interface{}) {
	if level <= l.level {
		m := fmt.Sprintf(msg, args...)
		if l.msgid > 0 {
			l.logger.Printf("[%s:%7d] %s\n", level.LogString(), l.msgid, m)
		} else {
			l.logger.Printf("[%s:default] %s\n", level.LogString(), m)
		}
	}
}

func (l *std_logger_wrapper) MessageId(id uint32) Logger {
	return &std_logger_wrapper{logger: l.logger, level: l.level, msgid: id}
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

func (l *nil_logger) Custom(level LogLevel, msg string, args ...interface{}) {}

func (l *nil_logger) MessageId(id uint32) Logger {
	return l
}
