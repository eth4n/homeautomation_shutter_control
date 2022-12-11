package domain

import (
	"strings"
)

func GetTopicPrefix(d Entity) string {
	cfg := d.GetAppState().Configuration
	uID := d.GetUniqueId()
	if uID != "" {
		uID = uID + "/"
	}
	return cfg.ChannelPrefix + "/" + cfg.NodeId + "/" + d.GetRawId() + "/" + uID
}

func GetDiscoveryTopic(d Entity) string {
	cfg := d.GetAppState().Configuration
	uID := d.GetUniqueId()
	if uID != "" {
		uID = uID + "/"
	}
	return cfg.DiscoverChannel + "/" + d.GetRawId() + "/" + cfg.NodeId + "/" + uID + "config"
}

func GetTopic(d Entity, rawTopicString string) string {
	return GetTopicPrefix(d) + strings.TrimSuffix(rawTopicString, "_topic")
}

func GetAvailabilityTopic(cfg *CtrlConfig) string {
	return cfg.ChannelPrefix + "/" + cfg.NodeId + "/availability"
}
