package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	logrus "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	handlers string

	mqttBrokerURL         string
	mqttBrokerPort        int
	mqttTopicSubscription string
	mqttTopicRegex        string

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

	flag.StringVar(&handlers, "handlers", os.Getenv("HANDLERS"), "Comma separated Handler to activate (default mqtt,http)")
	if handlers == "" {
		handlers = "mqtt,http"
	}

	// MQTT Settings
	flag.StringVar(&mqttBrokerURL, "mqtt-url", os.Getenv("MQTT_URL"), "MQTT Broker URL")
	flag.IntVar(&mqttBrokerPort, "mqtt-port", 1883, "port of MQTT Broker")
	flag.StringVar(&mqttTopicSubscription, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "MQTT Topic to subscribe")
	flag.StringVar(&mqttTopicRegex, "mqtt-topic-regex", os.Getenv("MQTT_TOPIC_REGEX"), "MQTT Topic regex to extract deviceID (by default, replace '+' by '(.+)' in mqtt-topic). deviceID must be first match")
	if mqttTopicRegex == "" {
		tmp := strings.Replace(mqttTopicSubscription, "+", "(.+)", -1)
		m := regexp.MustCompile("^\$share/.[^/]+/")
		mqttTopicRegex = m.ReplaceAllString(tmp, "")
	}
	// assert regex compile
	_ = regexp.MustCompile(mqttTopicRegex)

	// KAFKA Settings
	flag.BoolVar(&kafkaEventHub, "kafka-eventhub", false, "Use kafka eventhub compatibility (default false)")
	flag.StringVar(&kafkaUsername, "kafka-username", os.Getenv("KAFKA_USER_USERNAME"), "Username for kafka broker")
	flag.StringVar(&kafkaPassword, "kafka-password", os.Getenv("KAFKA_USER_PASSWORD"), "Password for kafka broker")
	flag.StringVar(&kafkaBrokerURL, "kafka-brokers", os.Getenv("KAFKA_BROKER"), "Kafka brokers in comma separated value")
	flag.BoolVar(&kafkaVerbose, "kafka-verbose", false, "Kafka verbose logging")
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

	activeHandlers := strings.Split(handlers, ",")
	// =====================================================
	// HTTP
	// =====================================================
	if contains(activeHandlers, "http") {
		handleHTTPRequests()
	}

	// =====================================================
	//  MQTT
	// =====================================================
	if contains(activeHandlers, "mqtt") {
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
	}

	<-sigchan
	os.Exit(0)
}
