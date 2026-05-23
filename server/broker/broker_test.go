package broker_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/mock/gomock"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/broker"
	"mqtt-streaming-server/domain"
	mock_domain "mqtt-streaming-server/mocks"
)

// mockMessage implements mqtt.Message interface
type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool      { return false }
func (m *mockMessage) Qos() byte            { return 0 }
func (m *mockMessage) Retained() bool       { return false }
func (m *mockMessage) Topic() string        { return m.topic }
func (m *mockMessage) MessageID() uint16    { return 0 }
func (m *mockMessage) Payload() []byte      { return m.payload }
func (m *mockMessage) Ack()                 {}

// mockMQTTClient implements mqtt.Client interface
type mockMQTTClient struct{}

type mockToken struct {
	err error
}

func (t *mockToken) Wait() bool                  { return true }
func (t *mockToken) WaitTimeout(d time.Duration) bool { return true }
func (t *mockToken) Error() error                { return t.err }
func (t *mockToken) Done() <-chan struct{}       { return nil }

func (m *mockMQTTClient) IsConnected() bool    { return true }
func (m *mockMQTTClient) IsAutoReconnect() bool { return true }
func (m *mockMQTTClient) IsConnectionOpen() bool { return true }
func (m *mockMQTTClient) Connect() mqtt.Token  { return &mockToken{} }
func (m *mockMQTTClient) Disconnect(uint)      {}
func (m *mockMQTTClient) Publish(string, byte, bool, interface{}) mqtt.Token {
	return &mockToken{}
}
func (m *mockMQTTClient) Subscribe(string, byte, mqtt.MessageHandler) mqtt.Token {
	return &mockToken{}
}
func (m *mockMQTTClient) SubscribeMultiple(map[string]byte, mqtt.MessageHandler) mqtt.Token {
	return &mockToken{}
}
func (m *mockMQTTClient) Unsubscribe(...string) mqtt.Token { return &mockToken{} }
func (m *mockMQTTClient) AddRoute(string, mqtt.MessageHandler)        {}
func (m *mockMQTTClient) OptionsReader() mqtt.ClientOptionsReader    { return mqtt.ClientOptionsReader{} }

// createPNGImage creates a valid PNG image payload
func createPNGImage() []byte {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestBrokerHandler_HandlePhoto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validImage := createPNGImage()

	tests := []struct {
		name        string
		topic       string
		payload     []byte
		setupMocks  func(*mock_domain.MockPhotoRepository, *mock_domain.MockDeviceRepository)
		shouldError bool
	}{
		{
			name:    "simple topic (ssproject/images)",
			topic:   "ssproject/images",
			payload: validImage,
			setupMocks: func(mockPhoto *mock_domain.MockPhotoRepository, mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "camera_stream").
					Return(&domain.Device{DeviceID: "camera_stream", DeviceName: "Camera Stream"}, nil)
				mockPhoto.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:    "topic with device ID",
			topic:   "ssproject/images/device123",
			payload: validImage,
			setupMocks: func(mockPhoto *mock_domain.MockPhotoRepository, mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device123").
					Return(&domain.Device{DeviceID: "device123", DeviceName: "Device 123"}, nil)
				mockPhoto.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:    "unknown device auto-register",
			topic:   "ssproject/images/unknown_device",
			payload: validImage,
			setupMocks: func(mockPhoto *mock_domain.MockPhotoRepository, mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "unknown_device").
					Return(nil, mongo.ErrNoDocuments)
				mockDevice.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
				mockPhoto.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:        "invalid image payload",
			topic:       "ssproject/images/device456",
			payload:     []byte("invalid image data"),
			shouldError: true,
			setupMocks: func(mockPhoto *mock_domain.MockPhotoRepository, mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device456").
					Return(&domain.Device{DeviceID: "device456", DeviceName: "Device 456"}, nil)
			},
		},
		{
			name:        "device repository error",
			topic:       "ssproject/images/device789",
			payload:     validImage,
			shouldError: true,
			setupMocks: func(mockPhoto *mock_domain.MockPhotoRepository, mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device789").
					Return(nil, mongo.ErrClientDisconnected)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPhoto := mock_domain.NewMockPhotoRepository(ctrl)
			mockDevice := mock_domain.NewMockDeviceRepository(ctrl)

			tt.setupMocks(mockPhoto, mockDevice)

			b := broker.NewBrokerHandlerWithRepos(mockPhoto, mockDevice)

			msg := &mockMessage{
				topic:   tt.topic,
				payload: tt.payload,
			}

			client := &mockMQTTClient{}
			b.HandlePhoto(client, msg)
		})
	}
}

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
