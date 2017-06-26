package gu_log

import (
	"fmt"
	"strconv"
	"strings"
)

type LogLevel uint

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

func (l LogLevel) LogString() string {
	switch l {
	case LOG_ERROR:
		return "ERR"
	case LOG_WARNING:
		return "WRN"
	case LOG_INFO:
		return "INF"
	case LOG_DEBUG:
		return "DBG"
	}
	return fmt.Sprintf("D%02d", l-LOG_DEBUG)
}
