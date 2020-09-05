package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	logrus "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	kafka "github.com/segmentio/kafka-go"
)

var (
	mqttBrokerURL         string
	mqttBrokerPort        int
	mqttTopicSubscription string

	kafkaBrokerURL     string
	kafkaVerbose       bool
	kafkaTopic         string
	kafkaConsumerGroup string
	kafkaClientID      string
	kafkaParitionCount int
)

// Telemetry : pool telemetry
type Telemetry struct {
	TemperatureIn  float32
	TemperatureOut float32
	Level          float32
}

// TelemetryDto : use to forward
type TelemetryDto struct {
	Unit  string
	Value float32
	Type  int
}

// TelemetryMetadata : pool telemetry with device
type TelemetryMetadata struct {
	DeviceID    string
	Telemetries []TelemetryDto
	Metadata    map[string]string
}

type logStruct struct {
	trace     string
	span      string
	component string
	logrus    logrus.Fields
}

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
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

func main() {
	flag.StringVar(&mqttBrokerURL, "mqtt-url", os.Getenv("MQTT_URL"), "MQTT Broker URL")
	flag.IntVar(&mqttBrokerPort, "mqtt-port", 1883, "port of MQTT Broker")
	flag.StringVar(&mqttTopicSubscription, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "MQTT Topic to subscribe")

	flag.StringVar(&kafkaBrokerURL, "kafka-brokers", os.Getenv("KAFKA_BROKER"), "Kafka brokers in comma separated value")
	flag.BoolVar(&kafkaVerbose, "kafka-verbose", true, "Kafka verbose logging")
	flag.StringVar(&kafkaTopic, "kafka-topic", os.Getenv("KAFKA_TOPIC"), "Kafka topic. Only one topic per worker.")
	flag.StringVar(&kafkaConsumerGroup, "kafka-consumer-group", os.Getenv("KAFKA_CONSUMER_GROUP"), "Kafka consumer group")
	flag.StringVar(&kafkaClientID, "kafka-client-id", os.Getenv("KAFKA_CLIENT_ID"), "Kafka client id")
	partCount, partCountErr := strconv.Atoi(os.Getenv("KAFKA_PARTITION_COUNT"))
	if partCountErr != nil {
		partCount = 5
	}
	flag.IntVar(&kafkaParitionCount, "kafka-partition-count", partCount, "# of Kafka partition")

	flag.Parse()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	var connString bytes.Buffer
	connString.WriteString(mqttBrokerURL)
	connString.WriteString(":")
	connString.WriteString(strconv.Itoa(mqttBrokerPort))

	// mqtt.DEBUG = log.New(os.Stdout, "MQTT [DEBUG]", 0)
	mqtt.ERROR = log.New(os.Stdout, "MQTT [ERROR]", 0)
	opts := mqtt.NewClientOptions().AddBroker(connString.String())
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	defer mqttClient.Unsubscribe(mqttTopicSubscription)
	defer mqttClient.Disconnect(250)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := mqttClient.Subscribe(mqttTopicSubscription, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	fmt.Println(fmt.Sprintf("Connected to MQTT Broker - Topic: %s", mqttTopicSubscription))

	<-sigchan
}

func decodeTelemetry(input []byte) (Telemetry, error) {
	reader := bytes.NewReader(input)
	t := Telemetry{}
	err := binary.Read(reader, binary.LittleEndian, &t)
	return t, err
}

func forwardTelemetry(device string, telemetry Telemetry, logStruct logStruct) {
	brokers := strings.Split(kafkaBrokerURL, ",")
	config := kafka.WriterConfig{
		MaxAttempts: 3,
		Brokers:     brokers,
		Topic:       kafkaTopic,
		Balancer:    &kafka.LeastBytes{},
	}

	writer := kafka.NewWriter(config)
	defer writer.Close()

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
		logrus.WithFields(logStruct.logrus).Error(err)
		return
	}

	partition := getPartition(device)
	fmt.Print(strconv.Itoa(partition))
	telemetryMessage := kafka.Message{
		Value: jsonTelemetry,
		Key:   []byte(strconv.Itoa(partition)),
	}

	err = writer.WriteMessages(context.Background(), telemetryMessage)
	if err != nil {
		logrus.WithFields(logStruct.logrus).Error(err)
	}
}

func getPartition(device string) (partition int) {
	sum := 0
	alphabet := "abcdefghijklmnopqrstuvwxyz1234567890"
	for _, letter := range device[0:5] {
		sum += strings.Index(alphabet, strings.ToLower(string(letter)))
	}
	fmt.Print(sum)
	return sum % kafkaParitionCount
}

func getRandomID(length int) (ID string) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
