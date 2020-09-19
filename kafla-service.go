package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	logrus "github.com/sirupsen/logrus"
)

func getKafkaCongif() *sarama.Config {
	config := sarama.NewConfig()
	config.Net.DialTimeout = 10 * time.Second
	config.Net.SASL.Enable = true
	if kafkaEventHub {
		config.Net.SASL.User = "$ConnectionString"
	} else {
		config.Net.SASL.User = kafkaUsername
	}
	config.Net.SASL.Password = kafkaPassword
	config.Net.SASL.Mechanism = "PLAIN"
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: true,
		ClientAuth:         0,
	}
	config.Version = sarama.V1_0_0_0
	config.Producer.Return.Successes = true
	return config
}

func forwardTelemetry(device string, telemetry Telemetry, logStruct logStruct) {
	brokers := strings.Split(kafkaBrokerURL, ",")
	producer, err := sarama.NewSyncProducer(brokers, getKafkaCongif())
	if err != nil {
		fmt.Println("Failed to start Sarama producer:", err)
		os.Exit(1)
	}

	telemetryWithMeta := TelemetryMetadata{
		DeviceID: device,
		Telemetries: []TelemetryDto{
			{Type: 1, Value: telemetry.TemperatureIn, Unit: "celsius"},
			{Type: 2, Value: telemetry.TemperatureOut, Unit: "celsius"},
			{Type: 3, Value: telemetry.Level, Unit: "m3"},
		},
		Metadata: map[string]string{
			"x-trace":     logStruct.trace,
			"x-span":      logStruct.span,
			"x-component": logStruct.component,
		},
	}

	jsonTelemetry, err := json.Marshal(telemetryWithMeta)
	if err != nil {
		logrus.WithFields(logStruct.logrus).Errorln(fmt.Sprintf("Unable to marshall data - %s", err))
		return
	}

	ts := time.Now().String()
	msg := &sarama.ProducerMessage{Topic: kafkaTopic, Key: sarama.StringEncoder("key-" + ts), Value: sarama.StringEncoder(jsonTelemetry)}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		logrus.WithFields(logStruct.logrus).Errorln(fmt.Sprintf("Fail to send message - %s", err))
	}
}
