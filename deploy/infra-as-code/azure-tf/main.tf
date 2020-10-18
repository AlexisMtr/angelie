resource "azurerm_resource_group" "angelie" {
  name     = var.rg_name
  location = var.rg_location
}

resource "azurerm_app_service_plan" "angelie" {
  name                = "${var.release_name}-angelie-service-plan"
  location            = "${azurerm_resource_group.angelie.location}"
  resource_group_name = "${azurerm_resource_group.angelie.name}"
  reserved            = true
  kind                = "Linux"

  sku {
    tier = var.sku_tier
    size = var.sku_size
  }
}

resource "azurerm_app_service" "angelie" {
  name                = "${var.release_name}-webapp"
  location            = "${azurerm_resource_group.angelie.location}"
  resource_group_name = "${azurerm_resource_group.angelie.name}"
  app_service_plan_id = "${azurerm_app_service_plan.angelie.id}"

  # Do not attach Storage by default
  app_settings {
    WEBSITES_ENABLE_APP_SERVICE_STORAGE = false
    APPINSIGHTS_INSTRUMENTATIONKEY      = var.create_app_insights ? azurerm_application_insights.angelie[0].instrumentation_key : var.app_insight_instrumentation_key

    # Settings for private Container Registires  
    DOCKER_REGISTRY_SERVER_URL      = var.registry_url
    DOCKER_REGISTRY_SERVER_USERNAME = var.registry_user
    DOCKER_REGISTRY_SERVER_PASSWORD = var.registry_password

    MQTT_URL              = var.mqtt_subscribe_connection_string
    MQTT_TOPIC            = var.mqtt_subscribe_topic
    MQTT_TOPIC_REGEX      = var.mqtt_subscribe_topic_regex
    KAFKA_BROKER          = var.event_publish_broker
    KAFKA_USER_PASSWORD   = var.event_publish_password
    KAFKA_TOPIC           = var.event_publish_topic
    KAFKA_PARTITION_COUNT = var.event_publish_part_count
  }

  # Configure Docker Image to load on start
  site_config {
    linux_fx_version = "DOCKER|${var.angelie_image}:${var.angelie_tag}"
    always_on        = "true"
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_application_insights" "angelie" {
  count = var.create_app_insights ? 1 : 0

  name                = "${var.release_name}-appinsights"
  location            = azurerm_resource_group.angelie.location
  resource_group_name = azurerm_resource_group.angelie.name
  application_type    = "web"
}