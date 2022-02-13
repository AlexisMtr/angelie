package main

import "math/rand"

func getRandomID(length int) (ID string) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func (telemetry Telemetry) MapToDto(device string, trace string, span string, component string) TelemetryMetadata {
	return TelemetryMetadata{
		DeviceID: device,
		Telemetries: []TelemetryDto{
			{Type: 0, Value: telemetry.TemperatureIn, Unit: "°C"},
			{Type: 1, Value: float32(telemetry.Ph), Unit: ""},
			{Type: 2, Value: telemetry.Level, Unit: "m3"},
			{Type: 3, Value: telemetry.Battery, Unit: "%"},
			{Type: 4, Value: telemetry.TemperatureOut, Unit: "°C"},
		},
		Metadata: map[string]string{
			"x-trace":     trace,
			"x-span":      span,
			"x-component": component,
		},
	}
}
