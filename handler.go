package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	logrus "github.com/sirupsen/logrus"
)

type daprSubscribeResponse struct {
	pubsubname string
	topic      string
	route      string
}

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

	pattern := regexp.MustCompile(mqttTopicRegex)
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

func handleHTTPRequests(activeHandlers []string) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/healthz", healthPage)

	if contains(activeHandlers, "http") {
		myRouter.HandleFunc("/telemetries/{deviceID}", httpTelemetryHandler).Methods("POST")
	}
	if contains(activeHandlers, "mqtt") {
		myRouter.HandleFunc("/dapr/subscribe", daprMqttHandler).Methods("GET")
		myRouter.HandleFunc("/mqtt/subscritpion", nil).Methods("POST")
	}

	go func() {
		// run in goroutine to avoid blocking
		log.Fatal(http.ListenAndServe(":10000", myRouter))
	}()
}

func healthPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func daprMqttHandler(w http.ResponseWriter, r *http.Request) {
	daprResp := daprSubscribeResponse{pubsubname: daprMqttCompName, route: "/mqtt/subscription", topic: mqttTopic}
	daprJSON, _ := json.Marshal(daprResp)
	w.WriteHeader(http.StatusOK)
	w.Write(daprJSON)
}

func httpTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]

	fmt.Println(reqBody)

	trace := getRandomID(10)
	span := getRandomID(10)
	component := "angelie"

	stdFields := logrus.Fields{
		"x-trace":     trace,
		"x-span":      span,
		"x-component": "angelie",
		"env":         os.Getenv("ENV"),
	}

	telemetry := Telemetry{}
	err := json.Unmarshal(reqBody, &telemetry)
	if err != nil {
		logrus.WithFields(stdFields).Errorln(fmt.Sprintf("Unable to decode playload %s - %s", reqBody, err))
		return
	}

	forwardTelemetry(string(deviceID), telemetry, logStruct{
		trace:     trace,
		span:      span,
		component: component,
		logrus:    stdFields,
	})
}
