package domain

// see https://github.com/W-Floyd/ha-mqtt-iot/blob/main/devices/externaldevice/switch.go

type Device struct {
	ConfigurationUrl string `json:"configuration_url,omitempty"` // "A link to the webpage that can manage the configuration of this types. Can be either an HTTP or HTTPS link."
	Connections      string `json:"connections,omitempty"`       // "A list of connections of the types to the outside world as a list of tuples `[connection_type, connection_identifier]`. For example the MAC address of a network interface: `\"connections\": [[\"mac\", \"02:5b:26:a8:dc:12\"]]`."
	Identifiers      string `json:"identifiers,omitempty"`       // "A list of IDs that uniquely identify the types. For example a serial number."
	Manufacturer     string `json:"manufacturer,omitempty"`      // "The manufacturer of the types."
	Model            string `json:"model,omitempty"`             // "The model of the types."
	Name             string `json:"name,omitempty"`              // "The name of the types."
	SuggestedArea    string `json:"suggested_area,omitempty"`    // "Suggest an area if the types isn’t in one yet."
	SwVersion        string `json:"sw_version,omitempty"`        // "The firmware version of the types."
	ViaDevice        string `json:"via_device,omitempty"`        // "Identifier of a types that routes messages between this types and Home Assistant. Examples of such devices are hubs, or parent devices of a sub-types. This is used to show types topology in Home Assistant."
}
