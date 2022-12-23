package main

import (
	"encoding/json"
	"fmt"
	"math"
	"shutter_control/common"
	"shutter_control/domain"
	"strconv"
	"time"
)

func getContactSensorValue(sensor *domain.BinarySensor) bool {
	if sensor == nil {
		return true
	}
	s := sensor.State
	if s != nil {
		var os domain.AqaraDoorSensorState
		json.Unmarshal([]byte(*s), &os)
		return *(&os.Contact)
	}
	return true
}
func calculateWindowValue(window *domain.StateWindow) {
	windowOpen := !getContactSensorValue(window.WindowOpenInputSensor)
	windowTilted := !getContactSensorValue(window.WindowTiltedInputSensor)
	scheduledPosition, _ := strconv.Atoi(*window.ScheduledValue.State)
	rainValue := state.RainInput.State

	openAndClosed := 100
	tiltedAndClosed := window.Config.TiltedAndClosed
	if windowOpen && scheduledPosition < openAndClosed {
		window.WindowOpenValue.UpdateState(String(strconv.Itoa(openAndClosed)))
	} else if !windowOpen && windowTilted && scheduledPosition < tiltedAndClosed {
		window.WindowOpenValue.UpdateState(String(strconv.Itoa(tiltedAndClosed)))
	} else if !windowOpen && !windowTilted {
		window.WindowOpenValue.UpdateState(String(""))
	}

	if windowOpen {
		window.WindowOpenState.UpdateState(String("2"))
	} else if !windowOpen && windowTilted {
		window.WindowOpenState.UpdateState(String("1"))
	} else if !windowOpen && !windowTilted {
		window.WindowOpenState.UpdateState(String("0"))
	}

	openAndDrizzle := window.Config.OpenAndDrizzle
	openAndStorm := window.Config.OpenAndStorm
	tiltedAndDrizzle := window.Config.TiltedAndDrizzle
	tiltedAndStorm := window.Config.TiltedAndStorm

	if windowOpen && *rainValue == domain.RainDrizzle && scheduledPosition > openAndDrizzle {
		window.RainValue.UpdateState(String(strconv.Itoa(openAndDrizzle)))
	} else if windowOpen && *rainValue == domain.RainStorm && scheduledPosition > openAndStorm {
		window.RainValue.UpdateState(String(strconv.Itoa(openAndStorm)))
	} else if !windowOpen && windowTilted && *rainValue == domain.RainDrizzle && scheduledPosition > tiltedAndDrizzle {
		window.RainValue.UpdateState(String(strconv.Itoa(tiltedAndDrizzle)))
	} else if !windowOpen && windowTilted && *rainValue == domain.RainStorm && scheduledPosition > tiltedAndStorm {
		window.RainValue.UpdateState(String(strconv.Itoa(tiltedAndStorm)))
	} else {
		window.RainValue.UpdateState(String(""))
	}
}
func windowOpenStateChanged(sensor *domain.BinarySensor, newState *bool, oldState *bool) {
	window := sensor.Window
	calculateWindowValue(window)
	recalculateWindow(sensor.Window)
}

func windowTiltedStateChanged(sensor *domain.BinarySensor, newState *bool, oldState *bool) {
	window := sensor.Window
	calculateWindowValue(window)
	recalculateWindow(sensor.Window)
}

func manualCoverStateChanged(cover *domain.Cover, newPosition int) {
	currentPosition := getCoverPosition(cover.Window.OutputCover)
	common.LogDebug(fmt.Sprintf("Manual cover %s set to position %d", *cover.UniqueId, newPosition))

	if newPosition == 100 && currentPosition < 99 {
		common.LogDebug(fmt.Sprintf("Fix manual cover %s to 99 instead of 100 (current position is %d)", *cover.UniqueId, currentPosition))
		newPosition = 99
	}

	position := strconv.Itoa(newPosition)
	cover.Window.ManualValue.UpdateState(&position)
	recalculateWindow(cover.Window)
}

func scheduledCoverStateChanged(cover *domain.Cover, newState *domain.CoverState, oldState *domain.CoverState) {
	window := cover.Window
	position := strconv.Itoa(*newState.Position)
	if *window.Automation.State == "ON" {
		common.LogDebug(fmt.Sprintf("Scheduled input %s changed, resetting manual value", *cover.UniqueId))
		window.ManualValue.UpdateState(String(""))
	}
	window.ScheduledValue.UpdateState(&position)
	calculateWindowValue(window)
	recalculateWindow(window)
}

func rainInputStateChanged(rainValue *domain.Select, newState *string) {

	for _, w := range rainValue.AppState.Windows {
		calculateWindowValue(&w)
		recalculateWindow(&w)
	}
}

func recalculateWindow(window *domain.StateWindow) {
	var automationValue int
	scheduledPosition, e := strconv.Atoi(*window.ScheduledValue.State)
	if e != nil {
		scheduledPosition = 0
	}
	windowOpenPosition, e := strconv.Atoi(*window.WindowOpenValue.State)
	if e != nil {
		windowOpenPosition = -1
	}
	rainPosition, e := strconv.Atoi(*window.RainValue.State)
	if e != nil {
		rainPosition = -1
	}
	manualPosition, e := strconv.Atoi(*window.ManualValue.State)
	if e != nil {
		manualPosition = -1
	}

	automationValue = scheduledPosition
	if windowOpenPosition >= 0 {
		automationValue = windowOpenPosition
	}
	if rainPosition != -1 {
		automationValue = rainPosition
	}

	if manualPosition != -1 {
		automationValue = manualPosition
	}

	automationValueS := strconv.Itoa(automationValue)
	window.OutputValue.UpdateState(&automationValueS)

	if *window.Automation.State == "ON" {
		// Tell cover the automationValue
		updateCover(window, automationValue)
	} else {
		// Tell cover the manualPosition
		updateCover(window, manualPosition)
	}

}

func getCoverPosition(sensor *domain.Cover) int {
	if sensor == nil {
		return 0
	}
	s := sensor.State
	if s != nil {
		var os domain.CoverState
		json.Unmarshal([]byte(*s), &os)
		return **(&os.Position)
	}
	return 0
}
func Int(v int) *int { return &v }

type CoverStatePosition struct {
	Position *int `json:"position"`
}

type CoverStateOnly struct {
	State *string `json:"state"`
}

func updateCover(window *domain.StateWindow, value int) {
	currentPosition := getCoverPosition(window.OutputCover)
	var newState CoverStatePosition
	var newStateString string
	var valueToGo = value

	if value == -2 {
		s := CoverStateOnly{
			State: String("STOP"),
		}
		j, _ := json.Marshal(s)
		newStateString = string(j)
	} else if value == -1 {

		return
	} else if value == 100 && currentPosition != 100 {
		common.LogDebug(fmt.Sprintf("Fixing calibration time to set value to 100 for window %s/%s (output cover: %s)", window.Id, window.Config.Id, window.Config.OutputCoverStateTopic))

		// Only to reset CalibrationTime and thus setting the position to 0
		token := (*window.OutputCover.AppState.Mqtt).Publish(window.Config.OutputCoverStateTopic+"/set/calibration_time", 0, false, strconv.Itoa(window.Config.OutputCoverTimeUp+10))
		token.Wait()

		token = (*window.OutputCover.AppState.Mqtt).Publish(window.Config.OutputCoverStateTopic+"/set/calibration_time", 0, false, strconv.Itoa(window.Config.OutputCoverTimeUp))
		token.Wait()
		time.Sleep(1 * time.Second)

		s := CoverStateOnly{
			State: String("OPEN"),
		}
		j, _ := json.Marshal(s)
		newStateString = string(j)
	} else if value < currentPosition {
		factor := float64(window.Config.OutputCoverTimeUp) / float64(window.Config.OutputCoverTimeDown)

		wayToGo := 100 - value
		a := float64(wayToGo) * factor
		b := a - float64(wayToGo)

		valueToGo = int(math.Round(float64(value) + b))

		newState = CoverStatePosition{
			Position: Int(valueToGo),
		}

		j, _ := json.Marshal(newState)
		newStateString = string(j)
	} else {
		newState = CoverStatePosition{
			Position: Int(value),
		}
		j, _ := json.Marshal(newState)
		newStateString = string(j)
	}

	if currentPosition == valueToGo {
		common.LogDebug(fmt.Sprintf("Skipping main cover update %s, new value %d equals current position %d", window.OutputCover.GetUniqueId(), value, currentPosition))
		return
	}

	common.LogDebug(fmt.Sprintf("Updating main cover %s=%d (%s, current=%d)", window.OutputCover.GetUniqueId(), value, newStateString, currentPosition))

	window.OutputCover.WriteCommand(String(newStateString))
}
