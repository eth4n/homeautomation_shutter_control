package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"os/signal"
	"shutter_control/common"
	"syscall"
	"time"
)

var mqttClient mqtt.Client

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	config := loadConfig()
	loadState()

	mqttClient = connect(config)
	state.Configuration = &config
	state.Mqtt = &mqttClient
	initEntities()

	//	mqtt.DEBUG = common.DebugLog
	mqtt.WARN = common.WarnLog
	mqtt.ERROR = common.ErrorLog
	mqtt.CRITICAL = common.CriticalLog

	stateUpdateTicker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-stateUpdateTicker.C:
				writeState()
			}
		}
	}()

	common.LogDebug("Everything is set up")
	makeAvailable()
	<-done

	stateUpdateTicker.Stop()
	writeState()
	common.LogDebug("Shutter control stopped")
	makeUnAvailable()
	disconnect(mqttClient)
	time.Sleep(1 * time.Second)
}
