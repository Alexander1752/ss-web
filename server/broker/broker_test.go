package broker

import (
	"context"
	"errors"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
)

func TestBrokerHandler_RegisterDevice(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		db        *mongo.Database
		// Named input parameters for target function.
		msg mqtt.Message
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
