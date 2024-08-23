package main

import (
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTT struct {
	mqttClient mqtt.Client
	logger     *log.Logger
}

func NewMQTTClient(port int, broker string, username, password string) (*MQTT, error) {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))

	// Create a logger
	logger := log.New(os.Stdout, "MQTT: ", log.LstdFlags)

	// Set handlers
	options.OnConnect = func(client mqtt.Client) {
		logger.Println("Connected to MQTT broker")
	}
	options.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.Printf("Connection lost: %v\n", err)
	}
	options.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {
		logger.Println("Reconnecting to broker...")
	}

	// Set credentials if provided
	if username != "" && password != "" {
		options.SetUsername(username)
		options.SetPassword(password)
	}

	client := mqtt.NewClient(options)

	// Connect to the broker
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, err
	}

	return &MQTT{
		mqttClient: client,
		logger:     logger,
	}, nil
}

func (mq *MQTT) Publish(topic string, payload []byte) error {
	mq.logger.Printf("Publishing to topic %s\n", topic)
	token := mq.mqttClient.Publish(topic, 0, false, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		mq.logger.Printf("Error publishing message: %v\n", err)
		return err
	}
	return nil
}

func (mq *MQTT) Subscribe(topic string, data chan []byte) error {
	mq.logger.Printf("Subscribing to topic %s\n", topic)
	token := mq.mqttClient.Subscribe(topic, 0, func(client mqtt.Client, message mqtt.Message) {
		message.Ack()
		data <- message.Payload()
		mq.logger.Printf("Received message from topic %s: %s\n", message.Topic(), string(message.Payload()))
	})
	token.Wait()
	if err := token.Error(); err != nil {
		mq.logger.Printf("Error subscribing to topic: %v\n", err)
		return err
	}
	return nil
}

func (mq *MQTT) Disconnect() {
	mq.logger.Println("Disconnecting MQTT client")
	mq.mqttClient.Disconnect(250)
}
