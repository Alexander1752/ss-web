package utils

import (
	"crypto/tls"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MQTTClient mqtt.Client

func InitMQTT() error {
	clientCert, err := tls.LoadX509KeyPair("/certs/client.crt", "/certs/client.key")
	if err != nil {
		return fmt.Errorf("failed to load MQTT client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      GetCACertPool(),
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
		ServerName:   os.Getenv("MQTT_HOST"),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker("ssl://" + os.Getenv("MQTT_HOST") + ":8883")
	opts.SetClientID("go-api-" + os.Getenv("HOSTNAME"))
	opts.SetTLSConfig(tlsConfig)
	opts.SetAutoReconnect(true)

	MQTTClient = mqtt.NewClient(opts)
	if token := MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	fmt.Println("Connected to MQTT Broker!")
	return nil
}

func PublishToMQTT(topic string, payload []byte) error {
	if MQTTClient == nil || !MQTTClient.IsConnected() {
		return fmt.Errorf("MQTT client is not initialized or disconnected")
	}

	token := MQTTClient.Publish(topic, 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("MQTT publish error: %w", token.Error())
	}

	return nil
}
