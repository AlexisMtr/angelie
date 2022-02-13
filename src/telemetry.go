package main

// Telemetry : pool telemetry
type Telemetry struct {
	TemperatureIn  float32
	TemperatureOut float32
	Level          float32
	Battery        float32
	Ph             int32
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
