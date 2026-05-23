package broker_test

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/broker"
)

func TestBrokerHandler_RegisterDevice(t *testing.T) {
	tests := []struct {
		name string
		db   *mongo.Database
		msg  mqtt.Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := broker.NewBrokerHandler(tt.db)
			b.RegisterDevice(nil, tt.msg)
		})
	}
}
