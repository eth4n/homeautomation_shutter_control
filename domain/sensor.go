package domain

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iancoleman/strcase"
	"log"
	"shutter_control/common"
	"time"
)

// see https://github.com/W-Floyd/ha-mqtt-iot/blob/main/devices/externaldevice/binary_sensor.go

type Sensor struct {
	AvailabilityMode       *string                          `json:"availability_mode,omitempty"`     // "When `availability` is configured, this controls the conditions needed to set the entity to `available`. Valid entries are `all`, `any`, and `latest`. If set to `all`, `payload_available` must be received on all configured availability topics before the entity is marked as online. If set to `any`, `payload_available` must be received on at least one configured availability topic before the entity is marked as online. If set to `latest`, the last `payload_available` or `payload_not_available` received on any configured availability topic controls the availability."
	AvailabilityTemplate   *string                          `json:"availability_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract device's availability from the `availability_topic`. To determine the devices's availability result of this template will be compared to `payload_available` and `payload_not_available`."
	AvailabilityTopic      *string                          `json:"availability_topic,omitempty"`    // "The MQTT topic subscribed to receive birth and LWT messages from the MQTT device. If `availability` is not defined, the binary sensor will always be considered `available` and its state will be `on`, `off` or `unknown`. If `availability` is defined, the binary sensor will be considered as `unavailable` by default and the sensor's initial state will be `unavailable`. Must not be used together with `availability`."
	Device                 *Device                          `json:"device,omitempty"`
	DeviceClass            *string                          `json:"device_class,omitempty"`             // "Sets the [class of the device](/integrations/binary_sensor/#device-class), changing the device state and icon that is displayed on the frontend."
	EnabledByDefault       *bool                            `json:"enabled_by_default,omitempty"`       // "Flag which defines if the entity should be enabled when first added."
	Encoding               *string                          `json:"encoding,omitempty"`                 // "The encoding of the payloads received. Set to `\"\"` to disable decoding of incoming payload."
	EntityCategory         *string                          `json:"entity_category,omitempty"`          // "The [category](https://developers.home-assistant.io/docs/core/entity#generic-properties) of the entity."
	ExpireAfter            *int                             `json:"expire_after,omitempty"`             // "If set, it defines the number of seconds after the sensor's state expires, if it's not updated. After expiry, the sensor's state becomes `unavailable`. Default the sensors state never expires."
	ForceUpdate            *bool                            `json:"force_update,omitempty"`             // "Sends update events (which results in update of [state object](/docs/configuration/state_object/)'s `last_changed`) even if the sensor's state hasn't changed. Useful if you want to have meaningful value graphs in history or want to create an automation that triggers on every incoming state message (not only when the sensor's new state is different to the current one)."
	Icon                   *string                          `json:"icon,omitempty"`                     // "[Icon](/docs/configuration/customizing-devices/#icon) for the entity."
	JsonAttributesTemplate *string                          `json:"json_attributes_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract the JSON dictionary from messages received on the `json_attributes_topic`. Usage example can be found in [MQTT sensor](/integrations/sensor.mqtt/#json-attributes-template-configuration) documentation."
	JsonAttributesTopic    *string                          `json:"json_attributes_topic,omitempty"`    // "The MQTT topic subscribed to receive a JSON dictionary payload and then set as sensor attributes. Usage example can be found in [MQTT sensor](/integrations/sensor.mqtt/#json-attributes-topic-configuration) documentation."
	Name                   *string                          `json:"name,omitempty"`                     // "The name of the binary sensor."
	ObjectId               *string                          `json:"object_id,omitempty"`                // "Used instead of `name` for automatic generation of `entity_id`"
	OffDelay               *int                             `json:"off_delay,omitempty"`                // "For sensors that only send `on` state updates (like PIRs), this variable sets a delay in seconds after which the sensor's state will be updated back to `off`."
	PayloadAvailable       *string                          `json:"payload_available,omitempty"`        // "The string that represents the `online` state."
	PayloadNotAvailable    *string                          `json:"payload_not_available,omitempty"`    // "The string that represents the `offline` state."
	PayloadOff             *string                          `json:"payload_off,omitempty"`              // "The string that represents the `off` state. It will be compared to the message in the `state_topic` (see `value_template` for details)"
	PayloadOn              *string                          `json:"payload_on,omitempty"`               // "The string that represents the `on` state. It will be compared to the message in the `state_topic` (see `value_template` for details)"
	Qos                    *int                             `json:"qos,omitempty"`                      // "The maximum QoS level to be used when receiving messages."
	StateTopic             *string                          `json:"state_topic,omitempty"`              // "The MQTT topic subscribed to receive sensor's state."
	UniqueId               *string                          `json:"unique_id,omitempty"`                // "An ID that uniquely identifies this sensor. If two sensors have the same unique ID, Home Assistant will raise an exception."
	ValueTemplate          *string                          `json:"value_template,omitempty"`           // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) that returns a string to be compared to `payload_on`/`payload_off` or an empty string, in which case the MQTT message will be removed. Available variables: `entity_id`. Remove this option when 'payload_on' and 'payload_off' are sufficient to match your payloads (i.e no pre-processing of original message is required)."
	AppState               *State                           `json:"-"`
	State                  *string                          `json:"-"`
	StateUpdatedFunc       *func(*Sensor, *string, *string) `json:"-"`
	Window                 *StateWindow                     `json:"-"`
}

func (d *Sensor) GetRawId() string {
	return "sensor"
}

func (d *Sensor) GetUniqueId() string {
	return *d.UniqueId
}
func (d *Sensor) UpdateState(state *string) {
	if state != nil {
		d.State = state
		common.LogDebug(fmt.Sprintf("Set Sensor state %s=%s", *d.UniqueId, *d.State))

		token := (*d.AppState.Mqtt).Publish(*d.StateTopic, byte(*d.Qos), false, *d.State)
		token.Wait()
	} else {
		d.State = state
		common.LogDebug(fmt.Sprintf("Set Sensor state %s=nil", *d.UniqueId))

		token := (*d.AppState.Mqtt).Publish(*d.StateTopic, byte(*d.Qos), false, nil)
		token.Wait()
	}

}

func (d *Sensor) Subscribe() {
	c := *d.AppState.Mqtt
	message, err := json.Marshal(d)
	if err != nil {
		log.Fatal(err)
	}

	if d.StateTopic != nil {
		t := c.Subscribe(*d.StateTopic, 0, d.handleStateUpdate())
		t.Wait()
		if t.Error() != nil {
			log.Fatal(t.Error())
		}
	}

	token := c.Publish(GetDiscoveryTopic(d), 0, true, message)
	token.Wait()
	time.Sleep(common.HADiscoveryDelay)
}

func (d *Sensor) handleStateUpdate() func(client mqtt.Client, msg mqtt.Message) {

	return func(client mqtt.Client, msg mqtt.Message) {
		newState := string(msg.Payload())
		oldState := d.State

		if oldState == nil || newState != *oldState {
			d.State = &newState
			common.LogDebug(fmt.Sprintf("Sensor state %s=%s", *d.UniqueId, *d.State))
		}

		d.AppState.States[*d.UniqueId] = newState

		if d.StateUpdatedFunc != nil {
			(*d.StateUpdatedFunc)(d, oldState, &newState)
		}
	}

}
func (d *Sensor) UnSubscribe() {
	c := *d.AppState.Mqtt
	if d.StateTopic != nil {
		t := c.Unsubscribe(*d.StateTopic)
		t.Wait()
		if t.Error() != nil {
			log.Fatal(t.Error())
		}
	}
}
func (d *Sensor) Initialize() {
	if d.Qos == nil {
		d.Qos = new(int)
		*d.Qos = int(common.QoS)
	}
	if d.UniqueId == nil {
		d.UniqueId = new(string)
		*d.UniqueId = d.AppState.Configuration.NodeId + "_" + strcase.ToSnake(*d.Name)
	}

	if d.State == nil {
		d.State = new(string)
		d.State = String("")
	}
	d.PopulateTopics()
	if val, ok := d.AppState.States[*d.UniqueId]; ok {
		d.State = new(string)
		*d.State = val
	}
}
func (d *Sensor) PopulateTopics() {
	if d.StateTopic == nil {

		d.AvailabilityTopic = new(string)
		*d.AvailabilityTopic = GetAvailabilityTopic(d.AppState.Configuration)

		d.StateTopic = new(string)
		*d.StateTopic = GetTopic(d, "state_topic")
	}
}

func (d *Sensor) GetAppState() *State {
	return d.AppState
}

func (d *Sensor) SetAppState(appState *State) {
	d.AppState = appState
}
