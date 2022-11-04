package pkg

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Controllers   []string
	QueryInterval time.Duration

	MQTTHost        string
	MQTTPort        int
	MQTTClientID    string
	MQTTUsername    string
	MQTTPassword    string
	MQTTTopicPrefix string
}

func GetConfig() Config {
	cfg := Config{}

	return Config{
		Controllers:     cfg.GetControllerList(),
		QueryInterval:   cfg.GetQueryInterval(),
		MQTTHost:        cfg.GetMQTTHost(),
		MQTTPort:        cfg.GetMQTTPort(),
		MQTTClientID:    cfg.GetMQTTClientID(),
		MQTTUsername:    cfg.GetMQTTUsername(),
		MQTTPassword:    cfg.GetMQTTPassword(),
		MQTTTopicPrefix: cfg.GetMQTTTopicPrefix(),
	}
}

func (c Config) GetControllerList() []string {
	controllers := []string{}
	parts := strings.Split(os.Getenv("CONTROLLER_LIST"), ",")

	for _, controller := range parts {
		controller = strings.ReplaceAll(controller, " ", "")
		controllers = append(controllers, controller)
	}

	return controllers
}

func (c Config) GetQueryInterval() time.Duration {
	secondsString, ok := os.LookupEnv("QUERY_INTERVAL_SECONDS")
	if !ok {
		return time.Second
	}

	seconds, err := strconv.Atoi(secondsString)
	if err != nil {
		return time.Second
	}

	return time.Second * time.Duration(seconds)
}

func (c Config) GetMQTTHost() string {
	return os.Getenv("MQTT_HOST")
}

func (c Config) GetMQTTPort() int {
	tmp, exists := os.LookupEnv("MQTT_PORT")
	if !exists {
		return 1883
	}

	port, err := strconv.Atoi(tmp)
	if err != nil {
		return 1883
	}

	return port
}

func (c Config) GetMQTTClientID() string {
	clientID := os.Getenv("MQTT_CLIENT_ID")

	if clientID == "" {
		return "alfred-ha-fpp-mqtt"
	}

	return clientID
}

func (c Config) GetMQTTUsername() string {
	return os.Getenv("MQTT_USERNAME")
}

func (c Config) GetMQTTPassword() string {
	return os.Getenv("MQTT_PASSWORD")
}

func (c Config) GetMQTTTopicPrefix() string {
	topicPrefix, exists := os.LookupEnv("MQTT_TOPIC_PREFIX")

	if !exists {
		return "alfred/ha-fpp-mqtt"
	}

	return topicPrefix
}
