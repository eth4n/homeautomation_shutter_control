package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"shutter_control/common"
	"shutter_control/domain"
)

func connect(config domain.CtrlConfig) mqtt.Client {

	options := mqtt.NewClientOptions()
	options.AddBroker(config.MqttHost)
	options.SetClientID(config.NodeId)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	common.LogDebug(fmt.Sprintf("Connected to %s", config.MqttHost))
	token = client.Subscribe(config.ChannelPrefix, 1, messagePubHandler)
	token.Wait()

	common.LogDebug(fmt.Sprintf("Subscribed to control topic %s", config.ChannelPrefix))

	return client
}

func disconnect(client mqtt.Client) {
	client.Disconnect(250)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Message %s received on topic %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}
