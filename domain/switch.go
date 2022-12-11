package domain

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	strcase "github.com/iancoleman/strcase"
	"log"
	"shutter_control/common"
	"time"
)

// see https://github.com/W-Floyd/ha-mqtt-iot/blob/main/devices/externaldevice/switch.go

type Switch struct {
	AvailabilityMode       *string                          `json:"availability_mode,omitempty"`     // "When `availability` is configured, this controls the conditions needed to set the entity to `available`. Valid entries are `all`, `any`, and `latest`. If set to `all`, `payload_available` must be received on all configured availability topics before the entity is marked as online. If set to `any`, `payload_available` must be received on at least one configured availability topic before the entity is marked as online. If set to `latest`, the last `payload_available` or `payload_not_available` received on any configured availability topic controls the availability."
	AvailabilityTemplate   *string                          `json:"availability_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract types's availability from the `availability_topic`. To determine the devices's availability result of this template will be compared to `payload_available` and `payload_not_available`."
	AvailabilityTopic      *string                          `json:"availability_topic,omitempty"`    // "The MQTT topic subscribed to receive availability (online/offline) updates. Must not be used together with `availability`."
	CommandTopic           *string                          `json:"command_topic,omitempty"`         // "The MQTT topic to publish commands to change the switch state."
	CommandFunc            mqtt.MessageHandler              `json:"-"`
	Device                 *Device                          `json:"device,omitempty"`
	DeviceClass            *string                          `json:"device_class,omitempty"`             // "The [type/class](/integrations/switch/#types-class) of the switch to set the icon in the frontend."
	EnabledByDefault       *bool                            `json:"enabled_by_default,omitempty"`       // "Flag which defines if the entity should be enabled when first added."
	Encoding               *string                          `json:"encoding,omitempty"`                 // "The encoding of the payloads received and published messages. Set to `\"\"` to disable decoding of incoming payload."
	EntityCategory         *string                          `json:"entity_category,omitempty"`          // "The [category](https://developers.home-assistant.io/docs/core/entity#generic-properties) of the entity."
	Icon                   *string                          `json:"icon,omitempty"`                     // "[Icon](/docs/configuration/customizing-devices/#icon) for the entity."
	JsonAttributesTemplate *string                          `json:"json_attributes_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract the JSON dictionary from messages received on the `json_attributes_topic`. Usage example can be found in [MQTT sensor](/integrations/sensor.mqtt/#json-attributes-template-configuration) documentation."
	JsonAttributesTopic    *string                          `json:"json_attributes_topic,omitempty"`    // "The MQTT topic subscribed to receive a JSON dictionary payload and then set as sensor attributes. Usage example can be found in [MQTT sensor](/integrations/sensor.mqtt/#json-attributes-topic-configuration) documentation."
	Name                   *string                          `json:"name,omitempty"`                     // "The name to use when displaying this switch."
	ObjectId               *string                          `json:"object_id,omitempty"`                // "Used instead of `name` for automatic generation of `entity_id`"
	Optimistic             *bool                            `json:"optimistic,omitempty"`               // "Flag that defines if switch works in optimistic mode."
	PayloadAvailable       *string                          `json:"payload_available,omitempty"`        // "The payload that represents the available state."
	PayloadNotAvailable    *string                          `json:"payload_not_available,omitempty"`    // "The payload that represents the unavailable state."
	PayloadOff             *string                          `json:"payload_off,omitempty"`              // "The payload that represents `off` state. If specified, will be used for both comparing to the value in the `state_topic` (see `value_template` and `state_off` for details) and sending as `off` command to the `command_topic`."
	PayloadOn              *string                          `json:"payload_on,omitempty"`               // "The payload that represents `on` state. If specified, will be used for both comparing to the value in the `state_topic` (see `value_template` and `state_on`  for details) and sending as `on` command to the `command_topic`."
	Qos                    *int                             `json:"qos,omitempty"`                      // "The maximum QoS level of the state topic. Default is 0 and will also be used to publishing messages."
	Retain                 *bool                            `json:"retain,omitempty"`                   // "If the published message should have the retain flag on or not."
	StateOff               *string                          `json:"state_off,omitempty"`                // "The payload that represents the `off` state. Used when value that represents `off` state in the `state_topic` is different from value that should be sent to the `command_topic` to turn the types `off`."
	StateOn                *string                          `json:"state_on,omitempty"`                 // "The payload that represents the `on` state. Used when value that represents `on` state in the `state_topic` is different from value that should be sent to the `command_topic` to turn the types `on`."
	StateTopic             *string                          `json:"state_topic,omitempty"`              // "The MQTT topic subscribed to receive state updates."
	State                  *string                          `json:"-"`
	UniqueId               *string                          `json:"unique_id,omitempty"`      // "An ID that uniquely identifies this switch types. If two switches have the same unique ID, Home Assistant will raise an exception."
	ValueTemplate          *string                          `json:"value_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract types's state from the `state_topic`. To determine the switches's state result of this template will be compared to `state_on` and `state_off`."
	AppState               *State                           `json:"-"`
	Window                 *StateWindow                     `json:"-"`
	StateUpdatedFunc       *func(*Switch, *string, *string) `json:"-"`
}

func (d *Switch) GetRawId() string {
	return "switch"
}

func (d *Switch) GetUniqueId() string {
	return *d.UniqueId
}
func (d *Switch) UpdateState(state *string) {
	if state != nil {
		d.State = state
	}

	common.LogDebug(fmt.Sprintf("Set switch state %s=%s", *d.UniqueId, *d.State))
	token := (*d.AppState.Mqtt).Publish(*d.StateTopic, byte(*d.Qos), *d.Retain, *d.State)
	token.Wait()
}

func (d *Switch) Subscribe() {
	c := *d.AppState.Mqtt
	message, err := json.Marshal(d)
	if err != nil {
		log.Fatal(err)
	}
	if d.CommandFunc != nil {
		t := c.Subscribe(*d.CommandTopic, 0, d.CommandFunc)
		t.Wait()
		if t.Error() != nil {
			log.Fatal(t.Error())
		}
		if d.Window != nil {
			d.AppState.Topics[*d.CommandTopic] = d.Window
		}

		token := c.Publish(GetDiscoveryTopic(d), 0, true, message)
		token.Wait()
		time.Sleep(common.HADiscoveryDelay)
		d.UpdateState(nil)
	}

	if d.StateTopic != nil {
		t := c.Subscribe(*d.StateTopic, 0, d.handleStateUpdate())
		t.Wait()
		if t.Error() != nil {
			log.Fatal(t.Error())
		}
	}
}

func (d *Switch) handleStateUpdate() func(client mqtt.Client, msg mqtt.Message) {

	return func(client mqtt.Client, msg mqtt.Message) {
		newState := string(msg.Payload())
		oldState := d.State

		if newState != *oldState {
			d.State = &newState
			common.LogDebug(fmt.Sprintf("Switch state %s=%s", *d.UniqueId, *d.State))
		}

		d.AppState.States[*d.UniqueId] = newState

		if d.StateUpdatedFunc != nil {
			(*d.StateUpdatedFunc)(d, &newState, oldState)
		}
	}

}
func (d *Switch) UnSubscribe() {
	c := *d.AppState.Mqtt
	if d.CommandTopic != nil {
		t := c.Unsubscribe(*d.CommandTopic)
		t.Wait()
		if t.Error() != nil {
			log.Fatal(t.Error())
		}
	}
	if d.StateTopic != nil {
		t := c.Unsubscribe(*d.StateTopic)
		t.Wait()
		if t.Error() != nil {
			log.Fatal(t.Error())
		}
	}
}
func (d *Switch) Initialize() {
	if d.Qos == nil {
		d.Qos = new(int)
		*d.Qos = int(common.QoS)
	}
	if d.Retain == nil {
		d.Retain = new(bool)
		*d.Retain = common.Retain
	}
	if d.UniqueId == nil {
		d.UniqueId = new(string)
		*d.UniqueId = d.AppState.Configuration.NodeId + "_" + strcase.ToSnake(*d.Name)

	}
	if d.State == nil {
		d.State = new(string)
		*d.State = "ON"
	}
	d.PopulateTopics()

	if val, ok := d.AppState.States[*d.UniqueId]; ok {
		d.State = new(string)
		*d.State = val
	}
}
func (d *Switch) PopulateTopics() {

	d.AvailabilityTopic = new(string)
	*d.AvailabilityTopic = GetAvailabilityTopic(d.AppState.Configuration)

	if d.CommandFunc != nil {
		d.CommandTopic = new(string)
		*d.CommandTopic = GetTopic(d, "command_topic")
	}

	d.StateTopic = new(string)
	*d.StateTopic = GetTopic(d, "state_topic")
}

func (d *Switch) GetAppState() *State {
	return d.AppState
}

func (d *Switch) SetAppState(appState *State) {
	d.AppState = appState
}
