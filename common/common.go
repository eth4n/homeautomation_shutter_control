package common

import (
	"log"
	"os"
	"time"
)

type LogLevels struct {
	Debug    bool
	Error    bool
	Warn     bool
	Critical bool
}

var (
	Retain           bool          = false
	QoS              byte          = 0
	HADiscoveryDelay time.Duration = 500 * time.Millisecond
	MachineID        string
	LogState         = LogLevels{
		Debug:    true,
		Error:    true,
		Warn:     true,
		Critical: true,
	}
)

var DebugLog = log.New(os.Stdout, "DEBUG   ", 0)
var ErrorLog = log.New(os.Stdout, "ERROR   ", 0)
var WarnLog = log.New(os.Stdout, "WARN    ", 0)
var CriticalLog = log.New(os.Stdout, "CRITICAL", 0)

const logPrefix = "[shutter-ctrl]  "

func LogError(message ...interface{}) {
	if LogState.Error {
		if len(message) > 1 {
			for _, mes := range message[:len(message)-2] {
				ErrorLog.Printf("%v%v\n", logPrefix, mes)
			}
		}
		ErrorLog.Fatalf("%v%v\n", logPrefix, message[len(message)-1])
	} else {
		os.Exit(1)
	}
}

func LogDebug(message ...interface{}) {
	currentTime := time.Now()
	if LogState.Debug {
		for _, mes := range message {
			DebugLog.Printf("%v%02d.%02d.%d %02d:%02d:%02d | %v\n", logPrefix, currentTime.Day(), currentTime.Month(), currentTime.Year(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), mes)
		}
	}
}

func LogWarning(message ...interface{}) {
	currentTime := time.Now()
	if LogState.Warn {
		for _, mes := range message {
			WarnLog.Printf("%v%02d.%02d.%d %02d:%02d:%02d | %v\n", logPrefix, currentTime.Day(), currentTime.Month(), currentTime.Year(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), mes)

		}
	}
}
