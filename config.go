package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"shutter_control/common"
	"shutter_control/domain"
)

func loadConfig() domain.CtrlConfig {
	// Open our jsonFile
	jsonFile, err := os.Open("config/configuration.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var cfg domain.CtrlConfig
	json.Unmarshal(byteValue, &cfg)
	j, _ := json.MarshalIndent(cfg, "", "\t")
	common.LogDebug(fmt.Sprintf("Configuration loaded successfully:  %s", string(j)))
	return cfg
}
