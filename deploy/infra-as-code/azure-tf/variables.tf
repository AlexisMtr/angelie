variable "rg_name" {
  type        = string
  description = "(optional) describe your variable"
}

variable "rg_location" {
  type        = string
  description = "(optional) describe your variable"
}

variable "release_name" {
  type        = string
  description = ""
}

variable "app_insight_instrumentation_key" {
  type        = string
  description = ""
}

variable "create_app_insights" {
  type    = bool
  default = false
}

variable "mqtt_subscribe_connection_string" {
  type        = string
  description = ""
}

variable "mqtt_subscribe_topic" {
  type        = string
  description = ""
}

variable "mqtt_subscribe_topic_regex" {
  type        = string
  description = ""
  default     = "devices/(.+)/telemetry$"
}

variable "event_publish_broker" {
  type        = string
  description = ""
}

variable "event_publish_password" {
  type        = string
  description = ""
}

variable "event_publish_topic" {
  type        = string
  description = ""
}

variable "event_publish_part_count" {
  type        = number
  description = ""
  default     = 2
}

variable "angelie_image" {
  type = string
}

variable "angelie_image_tag" {
  type    = string
  default = "latest"
}