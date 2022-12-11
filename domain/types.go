package domain

import mqtt "github.com/eclipse/paho.mqtt.golang"

type CtrlConfig struct {
	NodeId          string             `json:"id"`
	MqttHost        string             `json:"mqtt"`
	ChannelPrefix   string             `json:"channel"`
	DiscoverChannel string             `json:"homeassistant_discover"`
	Windows         []CtrlConfigWindow `json:"windows"`
}
type CtrlConfigWindow struct {
	Id                     string `json:"id"`
	TiltedSensorStateTopic string `json:"window_tilted_sensor"`
	WindowSensorStateTopic string `json:"window_open_sensor"`
	OutputCoverStateTopic  string `json:"cover_output"`
	OutputCoverTimeUp      int    `json:"cover_output_calibration_time_up"`
	OutputCoverTimeDown    int    `json:"cover_output_calibration_time_down"`
	OpenAndDrizzle         int    `json:"open_drizzle"`
	OpenAndStorm           int    `json:"open_storm"`
	TiltedAndDrizzle       int    `json:"tilted_drizzle"`
	TiltedAndStorm         int    `json:"tilted_storm"`
	TiltedAndClosed        int    `json:"tilted_closed"`
}

type CtrlState struct {
	states map[string]string
}

type State struct {
	Mqtt          *mqtt.Client
	Configuration *CtrlConfig
	RainInput     *Select
	Windows       []StateWindow
	Topics        map[string]*StateWindow
	States        map[string]string
}

type StateWindow struct {
	Id                      string
	Config                  *CtrlConfigWindow
	Automation              *Switch
	ScheduledInputCover     *Cover
	ScheduledValue          *Sensor
	WindowTiltedInputSensor *BinarySensor
	WindowOpenInputSensor   *BinarySensor
	WindowOpenValue         *Sensor
	WindowOpenState         *Sensor
	ManualInputCover        *Cover
	ManualValue             *Sensor
	RainValue               *Sensor
	OutputValue             *Sensor
	OutputCover             *Cover
}

type AqaraDoorSensorState struct {
	Contact bool `json:"contact"`
}
