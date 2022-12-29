package domain

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iancoleman/strcase"
	"log"
	"shutter_control/common"
	"strings"
	"time"
)

// see https://github.com/W-Floyd/ha-mqtt-iot/blob/main/devices/externaldevice/cover.go

type Cover struct {
	AvailabilityMode       *string                         `json:"availability_mode,omitempty"`     // "When `availability` is configured, this controls the conditions needed to set the entity to `available`. Valid entries are `all`, `any`, and `latest`. If set to `all`, `payload_available` must be received on all configured availability topics before the entity is marked as online. If set to `any`, `payload_available` must be received on at least one configured availability topic before the entity is marked as online. If set to `latest`, the last `payload_available` or `payload_not_available` received on any configured availability topic controls the availability."
	AvailabilityTemplate   *string                         `json:"availability_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract device's availability from the `availability_topic`. To determine the devices's availability result of this template will be compared to `payload_available` and `payload_not_available`."
	AvailabilityTopic      *string                         `json:"availability_topic,omitempty"`    // "The MQTT topic subscribed to to receive birth and LWT messages from the MQTT cover device. If an `availability` topic is not defined, the cover availability state will always be `available`. If an `availability` topic is defined, the cover availability state will be `unavailable` by default. Must not be used together with `availability`."
	CommandTopic           *string                         `json:"command_topic,omitempty"`         // "The MQTT topic to publish commands to control the cover."
	CommandFunc            mqtt.MessageHandler             `json:"-"`
	Device                 *Device                         `json:"device,omitempty"`
	DeviceClass            *string                         `json:"device_class,omitempty"`             // "Sets the [class of the device](/integrations/cover/), changing the device state and icon that is displayed on the frontend."
	EnabledByDefault       *bool                           `json:"enabled_by_default,omitempty"`       // "Flag which defines if the entity should be enabled when first added."
	Encoding               *string                         `json:"encoding,omitempty"`                 // "The encoding of the payloads received and published messages. Set to `\"\"` to disable decoding of incoming payload."
	EntityCategory         *string                         `json:"entity_category,omitempty"`          // "The [category](https://developers.home-assistant.io/docs/core/entity#generic-properties) of the entity."
	Icon                   *string                         `json:"icon,omitempty"`                     // "[Icon](/docs/configuration/customizing-devices/#icon) for the entity."
	JsonAttributesTemplate *string                         `json:"json_attributes_template,omitempty"` // "Defines a [template](/docs/configuration/templating/#using-templates-with-the-mqtt-integration) to extract the JSON dictionary from messages received on the `json_attributes_topic`. Usage example can be found in [MQTT sensor](/integrations/sensor.mqtt/#json-attributes-template-configuration) documentation."
	JsonAttributesTopic    *string                         `json:"json_attributes_topic,omitempty"`    // "The MQTT topic subscribed to receive a JSON dictionary payload and then set as sensor attributes. Usage example can be found in [MQTT sensor](/integrations/sensor.mqtt/#json-attributes-topic-configuration) documentation."
	Name                   *string                         `json:"name,omitempty"`                     // "The name of the cover."
	ObjectId               *string                         `json:"object_id,omitempty"`                // "Used instead of `name` for automatic generation of `entity_id`"
	Optimistic             *bool                           `json:"optimistic,omitempty"`               // "Flag that defines if switch works in optimistic mode."
	PayloadAvailable       *string                         `json:"payload_available,omitempty"`        // "The payload that represents the online state."
	PayloadClose           *string                         `json:"payload_close,omitempty"`            // "The command payload that closes the cover."
	PayloadNotAvailable    *string                         `json:"payload_not_available,omitempty"`    // "The payload that represents the offline state."
	PayloadOpen            *string                         `json:"payload_open,omitempty"`             // "The command payload that opens the cover."
	PayloadStop            *string                         `json:"payload_stop,omitempty"`             // "The command payload that stops the cover."
	PositionClosed         *int                            `json:"position_closed,omitempty"`          // "Number which represents closed position."
	PositionOpen           *int                            `json:"position_open,omitempty"`            // "Number which represents open position."
	PositionTemplate       *string                         `json:"position_template,omitempty"`        // "Defines a [template](/topics/templating/) that can be used to extract the payload for the `position_topic` topic. Within the template the following variables are available: `entity_id`, `position_open`; `position_closed`; `tilt_min`; `tilt_max`. The `entity_id` can be used to reference the entity's attributes with help of the [states](/docs/configuration/templating/#states) template function;"
	PositionTopic          *string                         `json:"position_topic,omitempty"`           // "The MQTT topic subscribed to receive cover position messages."
	Qos                    *int                            `json:"qos,omitempty"`                      // "The maximum QoS level to be used when receiving and publishing messages."
	Retain                 *bool                           `json:"retain,omitempty"`                   // "Defines if published messages should have the retain flag set."
	SetPositionTemplate    *string                         `json:"set_position_template,omitempty"`    // "Defines a [template](/topics/templating/) to define the position to be sent to the `set_position_topic` topic. Incoming position value is available for use in the template `{% raw %}{{ position }}{% endraw %}`. Within the template the following variables are available: `entity_id`, `position`, the target position in percent; `position_open`; `position_closed`; `tilt_min`; `tilt_max`. The `entity_id` can be used to reference the entity's attributes with help of the [states](/docs/configuration/templating/#states) template function;"
	SetPositionTopic       *string                         `json:"set_position_topic,omitempty"`       // "The MQTT topic to publish position commands to. You need to set position_topic as well if you want to use position topic. Use template if position topic wants different values than within range `position_closed` - `position_open`. If template is not defined and `position_closed != 100` and `position_open != 0` then proper position value is calculated from percentage position."
	StateClosed            *string                         `json:"state_closed,omitempty"`             // "The payload that represents the closed state."
	StateClosing           *string                         `json:"state_closing,omitempty"`            // "The payload that represents the closing state."
	StateOpen              *string                         `json:"state_open,omitempty"`               // "The payload that represents the open state."
	StateOpening           *string                         `json:"state_opening,omitempty"`            // "The payload that represents the opening state."
	StateStopped           *string                         `json:"state_stopped,omitempty"`            // "The payload that represents the stopped state (for covers that do not report `open`/`closed` state)."
	StateTopic             *string                         `json:"state_topic,omitempty"`              // "The MQTT topic subscribed to receive cover state messages. State topic can only read (`open`, `opening`, `closed`, `closing` or `stopped`) state."
	TiltClosedValue        *int                            `json:"tilt_closed_value,omitempty"`        // "The value that will be sent on a `close_cover_tilt` command."
	TiltCommandTemplate    *string                         `json:"tilt_command_template,omitempty"`    // "Defines a [template](/topics/templating/) that can be used to extract the payload for the `tilt_command_topic` topic. Within the template the following variables are available: `entity_id`, `tilt_position`, the target tilt position in percent; `position_open`; `position_closed`; `tilt_min`; `tilt_max`. The `entity_id` can be used to reference the entity's attributes with help of the [states](/docs/configuration/templating/#states) template function;"
	TiltCommandTopic       *string                         `json:"tilt_command_topic,omitempty"`       // "The MQTT topic to publish commands to control the cover tilt."
	TiltMax                *int                            `json:"tilt_max,omitempty"`                 // "The maximum tilt value."
	TiltMin                *int                            `json:"tilt_min,omitempty"`                 // "The minimum tilt value."
	TiltOpenedValue        *int                            `json:"tilt_opened_value,omitempty"`        // "The value that will be sent on an `open_cover_tilt` command."
	TiltOptimistic         *bool                           `json:"tilt_optimistic,omitempty"`          // "Flag that determines if tilt works in optimistic mode."
	TiltStatusTemplate     *string                         `json:"tilt_status_template,omitempty"`     // "Defines a [template](/topics/templating/) that can be used to extract the payload for the `tilt_status_topic` topic. Within the template the following variables are available: `entity_id`, `position_open`; `position_closed`; `tilt_min`; `tilt_max`. The `entity_id` can be used to reference the entity's attributes with help of the [states](/docs/configuration/templating/#states) template function;"
	TiltStatusTopic        *string                         `json:"tilt_status_topic,omitempty"`        // "The MQTT topic subscribed to receive tilt status update values."
	UniqueId               *string                         `json:"unique_id,omitempty"`                // "An ID that uniquely identifies this cover. If two covers have the same unique ID, Home Assistant will raise an exception."
	ValueTemplate          *string                         `json:"value_template,omitempty"`           // "Defines a [template](/topics/templating/) that can be used to extract the payload for the `state_topic` topic."
	AppState               *State                          `json:"-"`
	State                  *string                         `json:"-"`
	StateUpdatedFunc       *func(*Cover, *string, *string) `json:"-"`
	Window                 *StateWindow                    `json:"-"`
}

type CoverState struct {
	CalibrationTime *int    `json:"calibration_time"`
	Position        *int    `json:"position"`
	State           *string `json:"state"`
	Moving          *string `json:"moving"`
}

func (d *Cover) GetRawId() string {
	return "cover"
}

func (d *Cover) GetUniqueId() string {
	return *d.UniqueId
}
func (d *Cover) WriteCommand(state *string) {
	common.LogDebug(fmt.Sprintf("Writer cover command %s=%s", *d.UniqueId, *state))

	token := (*d.AppState.Mqtt).Publish(*d.CommandTopic, byte(*d.Qos), *d.Retain, *state)
	token.Wait()
}
func (d *Cover) UpdateState(state *string) {
	if state != nil {
		d.setState(state)
		common.LogDebug(fmt.Sprintf("Set cover state %s=%s", *d.UniqueId, *d.State))
	}
	if d.StateTopic != nil {
		token := (*d.AppState.Mqtt).Publish(*d.StateTopic, byte(*d.Qos), *d.Retain, *d.State)
		token.Wait()
	}
}

func (d *Cover) setState(state *string) {
	var so CoverState
	json.Unmarshal([]byte(*d.State), &so)

	if strings.HasPrefix(*state, "{") {
		var no CoverState
		json.Unmarshal([]byte(*state), &no)

		if no.CalibrationTime != nil {
			so.CalibrationTime = no.CalibrationTime
		}
		if no.Position != nil {
			so.Position = no.Position
		}
		if no.State != nil {
			so.State = no.State
		}
	} else {
		so.State = state
	}
	j, _ := json.Marshal(so)
	w := string(j)
	d.State = &w
}

func (d *Cover) Subscribe() {
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

func (d *Cover) handleStateUpdate() func(client mqtt.Client, msg mqtt.Message) {

	return func(client mqtt.Client, msg mqtt.Message) {
		newState := string(msg.Payload())
		oldState := d.State

		if newState != *oldState {
			d.setState(&newState)
			common.LogDebug(fmt.Sprintf("Cover state %s=%s", *d.UniqueId, *d.State))
		}

		d.AppState.States[*d.UniqueId] = newState

		if d.StateUpdatedFunc != nil {
			(*d.StateUpdatedFunc)(d, oldState, d.State)
		}
	}

}

func (d *Cover) UnSubscribe() {
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

func String(v string) *string { return &v }
func Int(v int) *int          { return &v }
func (d *Cover) Initialize(allowPositioning bool) {
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
		s := CoverState{
			Position: Int(0),
			State:    String("STOP"),
		}
		j, _ := json.MarshalIndent(s, "", "")
		d.State = new(string)
		*d.State = string(j)
	}
	d.PopulateTopics(allowPositioning)

	if val, ok := d.AppState.States[*d.UniqueId]; ok {
		d.State = new(string)
		*d.State = val
	}
}
func (d *Cover) PopulateTopics(allowPositioning bool) {

	d.AvailabilityTopic = new(string)
	*d.AvailabilityTopic = GetAvailabilityTopic(d.AppState.Configuration)

	if d.CommandFunc != nil {
		d.CommandTopic = new(string)
		*d.CommandTopic = GetTopic(d, "command_topic")
	}

	if allowPositioning {
		if d.StateTopic == nil {
			d.StateTopic = new(string)
			*d.StateTopic = GetTopic(d, "state_topic")
		}

		if d.JsonAttributesTopic == nil {
			d.JsonAttributesTopic = new(string)
			*d.JsonAttributesTopic = GetTopic(d, "state_topic")
		}

		if d.PositionTopic == nil {
			d.PositionTopic = new(string)
			*d.PositionTopic = GetTopic(d, "state_topic")
		}

		if d.PositionTemplate == nil {
			d.PositionTemplate = new(string)
			*d.PositionTemplate = "{{ value_json.position }}"
		}
		if d.SetPositionTemplate == nil {
			d.SetPositionTemplate = new(string)
			*d.SetPositionTemplate = "{ \"position\": {{ position }} }"
		}

		if d.SetPositionTopic == nil {
			d.SetPositionTopic = new(string)
			*d.SetPositionTopic = GetTopic(d, "command_topic")
		}

		if d.ValueTemplate == nil {
			d.ValueTemplate = new(string)
			*d.ValueTemplate = "{{ value_json.state }}"
		}
	}
}

func (d *Cover) GetAppState() *State {
	return d.AppState
}

func (d *Cover) SetAppState(appState *State) {
	d.AppState = appState
}
