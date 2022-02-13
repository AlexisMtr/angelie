package main

import (
	"flag"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logrus "github.com/sirupsen/logrus"
)

var (
	// Version : define angelie version
	Version string
)

var (
	mqttBrokerURL         string
	mqttBrokerPort        int
	mqttTopicSubscription string
	mqttTopicRegex        string

	daprPubTopic         string
	daprPubComponentName string
	daprPort             int
)

var mqttClient mqtt.Client

func main() {

	logrus.Infof("Angelie version: %s", Version)

	// MQTT Settings
	flag.StringVar(&mqttBrokerURL, "mqtt-url", os.Getenv("MQTT_URL"), "MQTT Broker URL")
	flag.IntVar(&mqttBrokerPort, "mqtt-port", 1883, "port of MQTT Broker")
	flag.StringVar(&mqttTopicSubscription, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "MQTT Topic to subscribe")
	flag.StringVar(&mqttTopicRegex, "mqtt-topic-regex", os.Getenv("MQTT_TOPIC_REGEX"), "MQTT Topic regex to extract deviceID (by default, replace '+' by '(.+)' in mqtt-topic). deviceID must be first match")
	if mqttTopicRegex == "" {
		tmp := strings.Replace(mqttTopicSubscription, "+", "(.+)", -1)
		m := regexp.MustCompile(`^\$share/.[^/]+/`)
		mqttTopicRegex = m.ReplaceAllString(tmp, "")
	}
	// assert regex compile
	_ = regexp.MustCompile(mqttTopicRegex)
	// DAPR settings
	flag.StringVar(&daprPubComponentName, "dapr-pub-component", os.Getenv("DAPR_PUB_COMPONENT"), "Publication Component Name")
	flag.StringVar(&daprPubTopic, "dapr-pub-topic", os.Getenv("DAPR_PUB_TOPIC"), "Publication Topic Name")
	flag.IntVar(&daprPort, "dapr-port", 3500, "Port used for DAPR")
	flag.Parse()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	HandleHTTPRequests()
	HandleMQTTRequests()

	defer disconnect()

	<-sigchan
	os.Exit(0)
}
