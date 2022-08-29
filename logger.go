package fctrl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	loggerLevelInfo  = "info"
	loggerLevelWarn  = "warn"
	loggerLevelError = "error"
	loggerLevelFatal = "fatal"
)

type ILoger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type DefaultLoger struct {
}

func newDefaultLoger() *DefaultLoger {
	log.SetFlags(0)
	return &DefaultLoger{}
}

func (l *DefaultLoger) Info(args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelInfo,
		"msg":   fmt.Sprint(args...),
	})
}

func (l *DefaultLoger) Warn(args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelWarn,
		"msg":   fmt.Sprint(args...),
	})
}

func (l *DefaultLoger) Error(args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelError,
		"msg":   fmt.Sprint(args...),
	})
}

func (l *DefaultLoger) Fatal(args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelFatal,
		"msg":   fmt.Sprint(args...),
	})
	os.Exit(0)
}

func (l *DefaultLoger) Infof(format string, args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelInfo,
		"msg":   fmt.Sprintf(format, args...),
	})
}
func (l *DefaultLoger) Warnf(format string, args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelWarn,
		"msg":   fmt.Sprintf(format, args...),
	})
}
func (l *DefaultLoger) Errorf(format string, args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelError,
		"msg":   fmt.Sprintf(format, args...),
	})
}
func (l *DefaultLoger) Fatalf(format string, args ...interface{}) {
	l.jsonPrint(map[string]interface{}{
		"level": loggerLevelFatal,
		"msg":   fmt.Sprintf(format, args...),
	})
	os.Exit(0)
}

func (l *DefaultLoger) jsonPrint(params map[string]interface{}) {
	params["mod"] = "fctrl"
	b, err := json.Marshal(params)
	if err != nil {
		return
	}
	log.Printf("%s", b)
}
