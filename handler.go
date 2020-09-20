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

func handleHTTPRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/healthz", healthPage)
	myRouter.HandleFunc("/telemetries/{deviceID}", httpTelemetryHandler).Methods("POST")
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	go func() {
		// run in goroutine to avoid blocking
		log.Fatal(http.ListenAndServe(":10000", myRouter))
	}()
}

func healthPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
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
	fmt.Println(telemetry, deviceID, component)
	// forwardTelemetry(string(deviceID), telemetry, logStruct{
	// 	trace:     trace,
	// 	span:      span,
	// 	component: component,
	// 	logrus:    stdFields,
	// })
}
