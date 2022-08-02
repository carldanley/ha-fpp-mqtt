package pkg

import (
	"encoding/json"
	"fmt"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

type MQTTClient struct {
	Log          *logrus.Entry
	client       MQTT.Client
	topicPrefix  string
	stateMachine *StateMachine
}

type LightMessage struct {
	Controller string `json:"controller,omitempty"`
	State      string `json:"state"`
	RGBColor   []int  `json:"color,omitempty"`
}

type FalconMessage struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func NewMQTTClient(log *logrus.Logger, host string, port int, clientID, username, password, topicPrefix string) *MQTTClient {
	mqttClient := MQTTClient{
		Log:         log.WithField("component", "mqttClient"),
		topicPrefix: topicPrefix,
	}

	mqttClient.client = mqttClient.getMQTTClient(host, port, clientID, username, password)
	return &mqttClient
}

func (mc *MQTTClient) getMQTTClient(host string, port int, clientID, username, password string) MQTT.Client {
	hostString := fmt.Sprintf("tcp://%s:%d", host, port)
	options := MQTT.NewClientOptions().AddBroker(hostString)
	options.SetClientID(clientID)
	options.SetUsername(username)
	options.SetPassword(password)
	options.SetAutoReconnect(true)
	options.SetCleanSession(true)
	options.OnConnect = mc.OnClientConnect
	options.OnConnectionLost = mc.OnClientDisconnect

	return MQTT.NewClient(options)
}

func (mc *MQTTClient) OnClientDisconnect(client MQTT.Client, err error) {
	mc.Log.Info("Disconnected from MQTT host")
}

func (mc *MQTTClient) OnClientConnect(client MQTT.Client) {
	mc.Log.Info("Connected to MQTT host")

	// cache this new client for later
	mc.client = client

	// listen for all of the incoming light commands
	token := mc.GetClient().Subscribe(mc.getTopic("#"), 0, mc.onMessage)
	if token.Wait() && token.Error() != nil {
		mc.Log.WithError(token.Error()).Errorf("Could not subscribe to %s", mc.topicPrefix)
	}
}

func (mc *MQTTClient) getTopic(postfix string) string {
	return fmt.Sprintf("%s/%s", mc.topicPrefix, postfix)
}

func (mc *MQTTClient) onMessage(client MQTT.Client, msg MQTT.Message) {
	topicParts := strings.Split(msg.Topic(), "/")

	if len(topicParts) < 4 {
		return
	}

	slug := topicParts[2]
	method := topicParts[3]
	payload := LightMessage{}

	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		return
	}

	switch strings.ToLower(method) {
	case "set":
		mc.HandleSetLightMessage(slug, payload)
	default:
		mc.Log.WithField("method", method).Debug("Skipping unknown message method")
	}
}

func (mc *MQTTClient) Start(stateMachine *StateMachine) error {
	mc.Log.Info("Starting up!")
	mc.stateMachine = stateMachine

	if token := mc.GetClient().Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (mc *MQTTClient) Stop() {
	mc.Log.Debug("Shutting down...")

	if mc.client.IsConnected() {
		mc.GetClient().Disconnect(1000)
	}

	mc.Log.Debug("Shutdown!")
}

func (mc *MQTTClient) GetClient() MQTT.Client {
	return mc.client
}

func (mc *MQTTClient) HandleSetLightMessage(slug string, payload LightMessage) {
	if payload.Controller == "" {
		mc.Log.Debug("Skipping set with no controller specified")
		return
	}

	currentState, exists := mc.stateMachine.OverlayModels.Load(slug)
	if !exists {
		mc.Log.Debug("Skipping set with no cache available")
		return
	}

	newState := "0"
	if strings.ToLower(payload.State) == "on" {
		newState = "1"
	}

	color := "#ffffff"
	if len(payload.RGBColor) == 3 {
		color = fmt.Sprintf("#%02x%02x%02x", payload.RGBColor[0], payload.RGBColor[1], payload.RGBColor[2])
	}

	topic := fmt.Sprintf("falcon/player/%s/set/command", payload.Controller)
	message := FalconMessage{
		Command: "Overlay Model Fill",
		Args: []string{
			currentState.(OverlayModel).Name,
			newState,
			color,
		},
	}

	mc.Log.Infof("Updating light: %s", currentState.(OverlayModel).Name)
	data, _ := json.Marshal(message)
	go mc.GetClient().Publish(topic, 0, false, data)

	boolState := false
	if newState == "1" {
		boolState = true
	}

	mc.stateMachine.UpdateBySlug(slug, "", boolState, payload.RGBColor)
}

func (mc *MQTTClient) PublishOverlayModelStatus(slug string, model OverlayModel) {
	state := "off"
	if model.State {
		state = "on"
	}

	payload := LightMessage{
		State: state,
	}

	if len(model.RGBColor) == 3 {
		payload.RGBColor = model.RGBColor
	}

	mc.Log.Infof("Publishing light update: %s", model.Name)
	data, _ := json.Marshal(payload)
	topic := mc.getTopic(fmt.Sprintf("%s/status", slug))
	go mc.GetClient().Publish(topic, 0, true, data)
}
