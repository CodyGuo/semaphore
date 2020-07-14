package util

import (
	"os"

	"github.com/CodyGuo/glog"
)

var log = glog.New(os.Stderr, glog.WithAutoCallDepth())

// LogWarning logs a warning with arbitrary field if error
func LogWarning(err error) {
	if err != nil {
		log.Warning(err)
	}
}

// LogWarningWithFields logs a warning with added field context if error
func LogWarningWithFields(err error, fileds glog.Fields) {
	if err != nil {
		log.Warningf("%+v, error: %v", fileds, err)
	}
}

// LogError logs an error with arbitrary field if error
func LogError(err error) {
	if err != nil {
		log.Error(err)
	}
}

// LogErrorWithFields logs a error with added field context if error
func LogErrorWithFields(err error, fileds glog.Fields) {
	if err != nil {
		log.Errorf("%+v, error: %v", fileds, err)
	}
}

// LogPanic logs and panics with arbitrary field if error
func LogPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// LogPanicWithFields logs and panics with added field context if error
func LogPanicWithFields(err error, fields glog.Fields) {
	if err != nil {
		log.Panicf("%+v, error: %v", fields, err)
	}
}
