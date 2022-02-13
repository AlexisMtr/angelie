package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func Forward(t *TelemetryMetadata, logCtx logrus.Fields) (err error) {
	logger := logrus.WithFields(logCtx)
	payload, err := json.Marshal(t)
	if err != nil {
		logger.Error("Unable to Marshal payload")
		return err
	}

	url := fmt.Sprintf("http://localhost:%d/v1.0/publish/%s/%s", daprPort, daprPubComponentName, daprPubTopic)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := http.Client{}
	res, err := client.Do(req)
	defer client.CloseIdleConnections()

	if err != nil {
		logger.Error(err)
		return err
	}
	defer res.Body.Close()

	return nil
}

func Handle(deviceID string, telemetryBytes []byte) (err error) {
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

	telemetry := Telemetry{}
	err = json.Unmarshal(telemetryBytes, &telemetry)
	if err != nil {
		logger.Errorln(fmt.Sprintf("Unable to decode playload %s - %s", telemetryBytes, err))
		return err
	}

	data := telemetry.MapToDto(string(deviceID), trace, span, component)
	Forward(&data, stdFields)

	return err
}

func HandleREST(w http.ResponseWriter, r *http.Request) {
	payload, _ := ioutil.ReadAll(r.Body)
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	err := Handle(deviceID, payload)
	if err != nil {
		w.WriteHeader(500)
		w.Write(payload)
		return
	}
	w.WriteHeader(200)
	w.Write(payload)
}

func HandleDAPR(w http.ResponseWriter, r *http.Request) {
	payload, _ := ioutil.ReadAll(r.Body)
	err := Handle("", payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func healthPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func HandleHTTPRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/healthz", healthPage)
	myRouter.HandleFunc("/telemetries/{deviceID}", HandleREST).Methods("POST")
	//myRouter.HandleFunc("/topic", HandleDAPR).Methods("POST")

	go func() {
		// run in goroutine to avoid blocking
		logrus.Info("Starting HTTP Server on port 10000")
		logrus.Fatal(http.ListenAndServe(":10000", myRouter))
	}()
}
