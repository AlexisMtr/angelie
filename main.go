package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	logrus "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
	kafkaEventHub      bool
	kafkaUsername      string
	kafkaPassword      string
)

type logStruct struct {
	trace     string
	span      string
	component string
	logrus    logrus.Fields
}

func main() {
	// MQTT Settings
	flag.StringVar(&mqttBrokerURL, "mqtt-url", os.Getenv("MQTT_URL"), "MQTT Broker URL")
	flag.IntVar(&mqttBrokerPort, "mqtt-port", 1883, "port of MQTT Broker (default 1883)")
	flag.StringVar(&mqttTopicSubscription, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "MQTT Topic to subscribe")

	// KAFKA Settings
	flag.BoolVar(&kafkaEventHub, "kafka-eventhub", false, "Use kafka eventhub compatibility (default false)")
	flag.StringVar(&kafkaUsername, "-kafka-username", os.Getenv("KAFKA_USER_USERNAME"), "Username for kafka broker")
	flag.StringVar(&kafkaPassword, "-kafka-password", os.Getenv("KAFKA_USER_PASSWORD"), "Password for kafka broker")
	flag.StringVar(&kafkaBrokerURL, "kafka-brokers", os.Getenv("KAFKA_BROKER"), "Kafka brokers in comma separated value")
	flag.BoolVar(&kafkaVerbose, "kafka-verbose", false, "Kafka verbose logging (default false)")
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

	var mqttConnString bytes.Buffer
	mqttConnString.WriteString(mqttBrokerURL)
	mqttConnString.WriteString(":")
	mqttConnString.WriteString(strconv.Itoa(mqttBrokerPort))

	// mqtt.DEBUG = log.New(os.Stdout, "MQTT [DEBUG]", 0)
	mqtt.ERROR = log.New(os.Stdout, "MQTT [ERROR]", 0)
	opts := mqtt.NewClientOptions().AddBroker(mqttConnString.String())
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(mqttMessageHandler)
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
