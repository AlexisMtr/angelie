package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logrus "github.com/sirupsen/logrus"
)

var mqttMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Println(msg.Payload())

	trace := getRandomID(10)
	span := getRandomID(10)
	component := "angelie"

	stdFields := logrus.Fields{
		"x-trace":     trace,
		"x-span":      span,
		"x-component": "angelie",
		"env":         os.Getenv("ENV"),
	}

	telemetry, err := decodeTelemetry(msg.Payload())
	if err != nil {
		logrus.WithFields(stdFields).Errorln(fmt.Sprintf("Unable to decode playload %s - %s", msg.Payload(), err))
		return
	}

	pattern := regexp.MustCompile("devices/(.*)/telemetry$")
	deviceID := pattern.FindSubmatch([]byte(msg.Topic()))[1]

	forwardTelemetry(string(deviceID), telemetry, logStruct{
		trace:     trace,
		span:      span,
		component: component,
		logrus:    stdFields,
	})
}

func decodeTelemetry(input []byte) (Telemetry, error) {
	reader := bytes.NewReader(input)
	t := Telemetry{}
	err := binary.Read(reader, binary.LittleEndian, &t)
	return t, err
}
