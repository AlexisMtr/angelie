package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logrus "github.com/sirupsen/logrus"
)

var mqttOnConnect mqtt.OnConnectHandler = func(c mqtt.Client) {
	logrus.Info("Connected")
}
var mqttOnLost mqtt.ConnectionLostHandler = func(c mqtt.Client, e error) {
	logrus.Info("Lost connection", e)
}

var mqttMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	trace := getRandomID(10)
	span := getRandomID(10)
	component := "angelie"

	stdFields := logrus.Fields{
		"x-trace":     trace,
		"x-span":      span,
		"x-component": component,
		"env":         os.Getenv("ENV"),
	}
	logger := logrus.WithFields(stdFields)

	logger.Info("TOPIC: ", msg.Topic())
	logger.Debug(msg.Payload())

	telemetry, err := decodeTelemetry(msg.Payload())
	if err != nil {
		logger.Errorf(fmt.Sprintf("Unable to decode playload %s - %s", msg.Payload(), err))
		return
	}

	pattern := regexp.MustCompile(mqttTopicRegex)
	deviceID := pattern.FindSubmatch([]byte(msg.Topic()))[1]

	data := telemetry.MapToDto(string(deviceID), trace, span, component)
	Forward(&data, stdFields)
}

func decodeTelemetry(input []byte) (Telemetry, error) {
	reader := bytes.NewReader(input)
	t := Telemetry{}
	err := binary.Read(reader, binary.LittleEndian, &t)
	return t, err
}

func HandleMQTTRequests() {
	var mqttConnString bytes.Buffer
	mqttConnString.WriteString(mqttBrokerURL)
	mqttConnString.WriteString(":")
	mqttConnString.WriteString(strconv.Itoa(mqttBrokerPort))

	opts := mqtt.NewClientOptions().AddBroker(mqttConnString.String())
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetOnConnectHandler(mqttOnConnect)
	opts.SetConnectionLostHandler(mqttOnLost)
	if hostname := os.Getenv("HOSTNAME"); hostname != "" {
		opts.SetClientID(hostname)
	}

	mqttClient := mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logrus.Error(token.Error())
		os.Exit(1)
	}

	if token := mqttClient.Subscribe(mqttTopicSubscription, 0, mqttMessageHandler); token.Wait() && token.Error() != nil {
		logrus.Error(token.Error())
		os.Exit(1)
	}

	logrus.Info("Connected to MQTT Broker - Topic ", mqttTopicSubscription)
}

func disconnect() {
	if mqttClient != nil {
		mqttClient.Unsubscribe(mqttTopicSubscription)
		mqttClient.Disconnect(250)
	}
}
