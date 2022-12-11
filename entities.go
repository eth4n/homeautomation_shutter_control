package main

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"shutter_control/common"
	"shutter_control/domain"
	"strconv"
	"strings"
)

func initEntities() {
	state.Topics = make(map[string]*domain.StateWindow)
	device := domain.Device{
		Identifiers:  state.Configuration.NodeId,
		Manufacturer: domain.Manufacturer,
		Model:        domain.SoftwareName,
		Name:         domain.InstanceName,
	}

	rainInputValues := []string{domain.RainNone, domain.RainDrizzle, domain.RainStorm}
	var rainInput = domain.Select{
		Device:      &device,
		Name:        String("rain_input"),
		CommandFunc: rainInputHandler,
		AppState:    &state,
		Options:     &rainInputValues,
		State:       String(domain.RainNone),
	}

	state.RainInput = &rainInput
	state.RainInput.Initialize()
	state.RainInput.Subscribe()

	initWindows()
}

func initWindows() {
	state.Windows = make([]domain.StateWindow, 0)
	for _, w := range state.Configuration.Windows {
		window := domain.Device{
			Identifiers:  state.Configuration.NodeId + "_" + w.Id,
			Manufacturer: domain.Manufacturer,
			Model:        domain.WindowName,
			Name:         "window_" + w.Id,
		}

		var scheduledValue = domain.Sensor{
			Device:   &window,
			Name:     String(w.Id + "_scheduled_value"),
			AppState: &state,
		}
		var windowOpenValue = domain.Sensor{
			Device:   &window,
			Name:     String(w.Id + "_window_open_value"),
			AppState: &state,
		}
		var windowOpenState = domain.Sensor{
			Device:   &window,
			Name:     String(w.Id + "_window_open_state"),
			AppState: &state,
		}
		var manualValue = domain.Sensor{
			Device:   &window,
			Name:     String(w.Id + "_manual_value"),
			AppState: &state,
		}
		var rainValue = domain.Sensor{
			Device:   &window,
			Name:     String(w.Id + "_rain_value"),
			AppState: &state,
		}
		var outputValue = domain.Sensor{
			Device:   &window,
			Name:     String(w.Id + "_automation_output"),
			AppState: &state,
		}

		var automation = domain.Switch{
			Device:      &window,
			Name:        String(w.Id + "_window_automation"),
			CommandFunc: windowAutomationSwitch,
			AppState:    &state,
		}
		var manualCover = domain.Cover{
			Device:              &window,
			Name:                String(w.Id + "_manual_cover"),
			CommandFunc:         windowManualCover,
			AppState:            &state,
			StateTopic:          &w.OutputCoverStateTopic,
			PositionTopic:       &w.OutputCoverStateTopic,
			JsonAttributesTopic: &w.OutputCoverStateTopic,
		}
		var scheduledCover = domain.Cover{
			Device:           &window,
			Name:             String(w.Id + "_scheduled_cover"),
			CommandFunc:      windowScheduledInput,
			AppState:         &state,
			StateUpdatedFunc: &scheduledCoverHandler,
		}

		var windowOpenSensor *domain.BinarySensor
		if w.WindowSensorStateTopic != "" {
			windowOpenSensor = &domain.BinarySensor{
				Name:             String(w.Id + "_window_open"),
				StateTopic:       &w.WindowSensorStateTopic,
				StateUpdatedFunc: &windowOpenHandler,
				AppState:         &state,
			}
		}

		var windowTiltedSensor *domain.BinarySensor
		if w.TiltedSensorStateTopic != "" {
			windowTiltedSensor = &domain.BinarySensor{
				Name:             String(w.Id + "_window_tilted"),
				StateTopic:       &w.TiltedSensorStateTopic,
				StateUpdatedFunc: &windowTiltedHandler,
				AppState:         &state,
			}
		}

		var outputCover = domain.Cover{
			Device:       &window,
			Name:         String(w.Id + "_output_cover"),
			AppState:     &state,
			StateTopic:   &w.OutputCoverStateTopic,
			CommandTopic: String(w.OutputCoverStateTopic + "/set"),
		}

		sw := domain.StateWindow{
			Id:                      w.Id,
			Config:                  &w,
			Automation:              &automation,
			ScheduledInputCover:     &scheduledCover,
			ScheduledValue:          &scheduledValue,
			ManualInputCover:        &manualCover,
			WindowOpenInputSensor:   windowOpenSensor,
			WindowTiltedInputSensor: windowTiltedSensor,
			WindowOpenValue:         &windowOpenValue,
			WindowOpenState:         &windowOpenState,
			ManualValue:             &manualValue,
			OutputValue:             &outputValue,
			OutputCover:             &outputCover,
			RainValue:               &rainValue,
		}
		automation.Window = &sw
		scheduledCover.Window = &sw
		scheduledValue.Window = &sw
		manualCover.Window = &sw
		windowOpenValue.Window = &sw
		windowOpenState.Window = &sw
		manualValue.Window = &sw
		outputValue.Window = &sw
		outputCover.Window = &sw
		rainValue.Window = &sw
		if windowOpenSensor != nil {
			windowOpenSensor.Window = &sw
			windowOpenSensor.Initialize()
			windowOpenSensor.Subscribe()
		}
		if windowTiltedSensor != nil {
			windowTiltedSensor.Window = &sw
			windowTiltedSensor.Initialize()
			windowTiltedSensor.Subscribe()
		}

		automation.Initialize()
		automation.Subscribe()

		scheduledCover.Initialize(true)
		scheduledCover.Subscribe()

		scheduledValue.Initialize()
		scheduledValue.Subscribe()

		manualCover.Initialize(true)
		manualCover.Subscribe()

		manualValue.Initialize()
		manualValue.Subscribe()

		windowOpenValue.Initialize()
		windowOpenValue.Subscribe()

		windowOpenState.Initialize()
		windowOpenState.Subscribe()

		outputValue.Initialize()
		outputValue.Subscribe()

		outputCover.Initialize(true)
		outputCover.Subscribe()

		rainValue.Initialize()
		rainValue.Subscribe()

		state.Windows = append(state.Windows, sw)
	}
}

var windowOpenHandler = func(sensor *domain.BinarySensor, oldState *string, newState *string) {
	var oo *bool
	var ns domain.AqaraDoorSensorState
	json.Unmarshal([]byte(*newState), &ns)
	if oldState != nil {
		var os domain.AqaraDoorSensorState
		json.Unmarshal([]byte(*oldState), &os)
		oo = &os.Contact
	}
	windowOpenStateChanged(sensor, &ns.Contact, oo)
}

var windowTiltedHandler = func(sensor *domain.BinarySensor, oldState *string, newState *string) {
	var oo *bool
	var ns domain.AqaraDoorSensorState
	json.Unmarshal([]byte(*newState), &ns)
	if oldState != nil {
		var os domain.AqaraDoorSensorState
		json.Unmarshal([]byte(*oldState), &os)
		oo = &os.Contact
	}
	windowTiltedStateChanged(sensor, &ns.Contact, oo)
}

var windowAutomationSwitch mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	window := state.Topics[msg.Topic()]
	value := string(msg.Payload())
	window.Automation.UpdateState(&value)
}

type CoverStateAndPosition struct {
	State    *string `json:"state"`
	Position *int    `json:"position"`
}

var windowManualCover mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	window := state.Topics[msg.Topic()]
	value := string(msg.Payload())

	if value == "OPEN" {
		manualCoverStateChanged(window.ManualInputCover, 100)
	} else if value == "CLOSE" {
		manualCoverStateChanged(window.ManualInputCover, 0)
	} else if value == "STOP" {
		manualCoverStateChanged(window.ManualInputCover, -2)
	} else if !strings.HasPrefix(value, "{") {
		p, _ := strconv.Atoi(value)
		manualCoverStateChanged(window.ManualInputCover, p)
	} else {
		var no CoverStateAndPosition
		json.Unmarshal([]byte(value), &no)
		manualCoverStateChanged(window.ManualInputCover, *no.Position)
	}
}

var windowScheduledInput mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	window := state.Topics[msg.Topic()]
	value := string(msg.Payload())

	if value == "OPEN" {
		s := CoverStateAndPosition{
			State:    String("OPEN"),
			Position: Int(100),
		}
		j, _ := json.Marshal(s)
		window.ScheduledInputCover.UpdateState(String(string(j)))
	} else if value == "CLOSE" {
		s := CoverStateAndPosition{
			State:    String("CLOSE"),
			Position: Int(0),
		}
		j, _ := json.Marshal(s)
		window.ScheduledInputCover.UpdateState(String(string(j)))
	} else {
		window.ScheduledInputCover.UpdateState(&value)
	}
}
var scheduledCoverHandler = func(cover *domain.Cover, oldState *string, newState *string) {
	var ns domain.CoverState
	var os domain.CoverState
	json.Unmarshal([]byte(*newState), &ns)
	json.Unmarshal([]byte(*oldState), &os)

	scheduledCoverStateChanged(cover, &ns, &os)
}

var rainInputHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())

	state.RainInput.UpdateState(&value)
	rainInputStateChanged(state.RainInput, &value)
}

func makeAvailable() {
	c := *state.Mqtt
	token := c.Publish(domain.GetAvailabilityTopic(state.Configuration), 0, true, "online")
	token.Wait()
	common.LogDebug("Now available")
}

func makeUnAvailable() {
	c := *state.Mqtt
	token := c.Publish(domain.GetAvailabilityTopic(state.Configuration), 0, true, "offline")
	token.Wait()
}

func String(v string) *string { return &v }
