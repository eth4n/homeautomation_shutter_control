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

// see https://github.com/W-Floyd/ha-mqtt-iot/blob/main/devices/externaldevice/select.go

type Select struct {
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
	Options                *([]string)                      `json:"options,omitempty"`                  // "List of options that can be selected. An empty list or a list with a single item is allowed."
	Qos                    *int                             `json:"qos,omitempty"`                      // "The maximum QoS level of the state topic. Default is 0 and will also be used to publishing messages."
	Retain                 *bool                            `json:"retain,omitempty"`                   // "If the published message should have the retain flag on or not."
	State                  *string                          `json:"-"`
	StateTopic             *string                          `json:"state_topic,omitempty"`    // "The MQTT topic subscribed to receive sensor's state."
	UniqueId               *string                          `json:"unique_id,omitempty"`      // "An ID that uniquely identifies this switch types. If two switches have the same unique ID, Home Assistant will raise an exception."
	ValueTemplate          *string                          `json:"value_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract types's state from the `state_topic`. To determine the switches's state result of this template will be compared to `state_on` and `state_off`."
	AppState               *State                           `json:"-"`
	Window                 *StateWindow                     `json:"-"`
	StateUpdatedFunc       *func(*Select, *string, *string) `json:"-"`
}

func (d *Select) GetRawId() string {
	return "select"
}

func (d *Select) GetUniqueId() string {
	return *d.UniqueId
}
func (d *Select) UpdateState(state *string) {
	if state != nil {
		d.State = state
	}

	common.LogDebug(fmt.Sprintf("Set select state %s=%s", *d.UniqueId, *d.State))
	token := (*d.AppState.Mqtt).Publish(*d.StateTopic, byte(*d.Qos), *d.Retain, *d.State)
	token.Wait()
}

func (d *Select) Subscribe() {
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

func (d *Select) handleStateUpdate() func(client mqtt.Client, msg mqtt.Message) {

	return func(client mqtt.Client, msg mqtt.Message) {
		newState := string(msg.Payload())
		oldState := d.State

		if newState != *oldState {
			d.State = &newState
			common.LogDebug(fmt.Sprintf("Select state %s=%s", *d.UniqueId, *d.State))
		}

		d.AppState.States[*d.UniqueId] = newState

		if d.StateUpdatedFunc != nil {
			(*d.StateUpdatedFunc)(d, &newState, oldState)
		}
	}

}
func (d *Select) UnSubscribe() {
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
func (d *Select) Initialize() {
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
	d.PopulateTopics()
	if val, ok := d.AppState.States[*d.UniqueId]; ok {
		d.State = new(string)
		*d.State = val
	}
}
func (d *Select) PopulateTopics() {

	d.AvailabilityTopic = new(string)
	*d.AvailabilityTopic = GetAvailabilityTopic(d.AppState.Configuration)

	if d.CommandFunc != nil {
		d.CommandTopic = new(string)
		*d.CommandTopic = GetTopic(d, "command_topic")
	}

	d.StateTopic = new(string)
	*d.StateTopic = GetTopic(d, "state_topic")
}

func (d *Select) GetAppState() *State {
	return d.AppState
}

func (d *Select) SetAppState(appState *State) {
	d.AppState = appState
}
