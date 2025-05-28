package data

import (
	"checkerbox/internal/test"

	"gorm.io/gorm"
)

type LogType int

const (
	INFO LogType = iota
	WARNING
	ERROR
)

func (lt LogType) String() string {
	return [...]string{"INFO", "WARNING", "ERROR"}[lt]
}

type Log struct {
	gorm.Model
	LogType string
	Source  string
	Site    int
	Message string
}

func NewCustomLog(source, message string, site int, logType LogType) *Log {
	return &Log{
		LogType: logType.String(),
		Source:  source,
		Message: message,
		Site:    site,
	}
}

func NewResultLog(source string, result test.Result) *Log {
	var logType LogType
	if result.Result == test.Error {
		logType = ERROR
	} else {
		logType = INFO
	}
	return &Log{
		LogType: logType.String(),
		Source:  source,
		Message: result.Result.String() + "|" + result.Label + "|" + result.Message,
		Site:    result.Site,
	}
}
