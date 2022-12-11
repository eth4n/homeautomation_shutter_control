package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"shutter_control/common"
	"shutter_control/domain"
	"time"
)

var state = domain.State{}

func loadState() {
	state.States = make(map[string]string)

	// Open our jsonFile
	jsonFile, err := os.Open("config/states.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println("config/states.json not readable")
		return
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &state.States)
	j, _ := json.MarshalIndent(state.States, "", "\t")
	common.LogDebug(fmt.Sprintf("States loaded successfully:  %s", string(j)))

}

func writeState() {
	currentTime := time.Now()
	state.States["time"] = fmt.Sprintf("%02d.%02d.%d %02d:%02d:%02d", currentTime.Day(), currentTime.Month(), currentTime.Year(), currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	file, _ := json.MarshalIndent(state.States, "", " ")

	_ = ioutil.WriteFile("config/states.json", file, 0644)
}
