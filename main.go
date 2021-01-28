package main

import (
	"flag"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	logrus "github.com/sirupsen/logrus"
)

var (
	daprPort int
	handlers string

	daprMqttCompName string
	mqttTopic        string
	mqttTopicRegex   string

	daprURL string
)

type logStruct struct {
	trace     string
	span      string
	component string
	logrus    logrus.Fields
}

func main() {
	_daprPort, err := strconv.Atoi(os.Getenv("DAPR_PORT"))
	if err != nil {
		_daprPort = 3500
	}
	flag.IntVar(&daprPort, "dapr-port", _daprPort, "")

	flag.StringVar(&handlers, "handlers", os.Getenv("HANDLERS"), "Comma separated Handler to activate (default mqtt,http)")
	if handlers == "" {
		handlers = "mqtt,http"
	}

	flag.StringVar(&mqttTopic, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "MQTT Topic to subscribe")
	flag.StringVar(&mqttTopicRegex, "mqtt-topic-regex", os.Getenv("MQTT_TOPIC_REGEX"), "MQTT Topic regex to extract deviceID (by default, replace '+' by '(.+)' in mqtt-topic). deviceID must be first match")
	if mqttTopicRegex == "" {
		tmp := strings.Replace(mqttTopic, "+", "(.+)", -1)
		m := regexp.MustCompile("^\\$share/.[^/]+/")
		mqttTopicRegex = m.ReplaceAllString(tmp, "")
	}
	// assert regex compile
	_ = regexp.MustCompile(mqttTopicRegex)

	flag.Parse()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	activeHandlers := strings.Split(handlers, ",")

	handleHTTPRequests(activeHandlers)

	<-sigchan
	os.Exit(0)
}
