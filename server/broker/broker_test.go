package broker_test

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"

	"mqtt-streaming-server/broker"
	"mqtt-streaming-server/domain"
	mock_domain "mqtt-streaming-server/mocks"
)

// mockMessage implements mqtt.Message interface
type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Qos() byte         { return 0 }
func (m *mockMessage) Retained() bool    { return false }
func (m *mockMessage) Topic() string     { return m.topic }
func (m *mockMessage) MessageID() uint16 { return 0 }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Ack()              {}

// mockMQTTClient implements mqtt.Client interface
type mockMQTTClient struct{}

type mockToken struct {
	err error
}

func (t *mockToken) Wait() bool                       { return true }
func (t *mockToken) WaitTimeout(d time.Duration) bool { return true }
func (t *mockToken) Error() error                     { return t.err }
func (t *mockToken) Done() <-chan struct{}            { return nil }

func (m *mockMQTTClient) IsConnected() bool      { return true }
func (m *mockMQTTClient) IsAutoReconnect() bool  { return true }
func (m *mockMQTTClient) IsConnectionOpen() bool { return true }
func (m *mockMQTTClient) Connect() mqtt.Token    { return &mockToken{} }
func (m *mockMQTTClient) Disconnect(uint)        {}
func (m *mockMQTTClient) Publish(string, byte, bool, interface{}) mqtt.Token {
	return &mockToken{}
}
func (m *mockMQTTClient) Subscribe(string, byte, mqtt.MessageHandler) mqtt.Token {
	return &mockToken{}
}
func (m *mockMQTTClient) SubscribeMultiple(map[string]byte, mqtt.MessageHandler) mqtt.Token {
	return &mockToken{}
}
func (m *mockMQTTClient) Unsubscribe(...string) mqtt.Token        { return &mockToken{} }
func (m *mockMQTTClient) AddRoute(string, mqtt.MessageHandler)    {}
func (m *mockMQTTClient) OptionsReader() mqtt.ClientOptionsReader { return mqtt.ClientOptionsReader{} }

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
		name       string
		topic      string
		payload    []byte
		setupMocks func(*mock_domain.MockDeviceRepository)
	}{
		{
			name:    "new device with json payload is saved",
			topic:   "register/device-1",
			payload: []byte(`{"name":"Camera 1","ip":"192.168.1.10","port":"1883"}`),
			setupMocks: func(mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device-1").
					Return(nil, mongo.ErrNoDocuments)
				mockDevice.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, device *domain.Device) error {
						if device.DeviceID != "device-1" {
							t.Fatalf("expected device_id device-1, got %q", device.DeviceID)
						}
						if device.DeviceName != "Camera 1" {
							t.Fatalf("expected device_name Camera 1, got %q", device.DeviceName)
						}
						if device.IPAddress != "192.168.1.10" {
							t.Fatalf("expected ip 192.168.1.10, got %q", device.IPAddress)
						}
						if device.Port != "1883" {
							t.Fatalf("expected port 1883, got %q", device.Port)
						}
						if device.DeviceStatus != "active" {
							t.Fatalf("expected active status, got %q", device.DeviceStatus)
						}
						if device.LastSeen.IsZero() {
							t.Fatalf("expected LastSeen to be set")
						}
						return nil
					})
			},
		},
		{
			name:    "new device with plain text payload uses payload as name",
			topic:   "register/device-2",
			payload: []byte("Camera Two"),
			setupMocks: func(mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device-2").
					Return(nil, mongo.ErrNoDocuments)
				mockDevice.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, device *domain.Device) error {
						if device.DeviceName != "Camera Two" {
							t.Fatalf("expected payload to become device name, got %q", device.DeviceName)
						}
						if device.IPAddress != "" || device.Port != "" {
							t.Fatalf("expected empty ip and port for plain payload, got ip=%q port=%q", device.IPAddress, device.Port)
						}
						return nil
					})
			},
		},
		{
			name:    "existing device is updated",
			topic:   "register/device-3",
			payload: []byte(`{"name":"Camera 3","ip":"10.0.0.3","port":"8883"}`),
			setupMocks: func(mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device-3").
					Return(&domain.Device{DeviceID: "device-3", DeviceName: "Old Camera"}, nil)
				mockDevice.EXPECT().
					Update(gomock.Any(), "device-3", gomock.Any()).
					DoAndReturn(func(_ any, _ string, device *domain.Device) error {
						if device.DeviceID != "device-3" {
							t.Fatalf("expected device_id device-3, got %q", device.DeviceID)
						}
						if device.DeviceName != "Camera 3" {
							t.Fatalf("expected updated name Camera 3, got %q", device.DeviceName)
						}
						if device.IPAddress != "10.0.0.3" || device.Port != "8883" {
							t.Fatalf("expected updated endpoint 10.0.0.3:8883, got %s:%s", device.IPAddress, device.Port)
						}
						if device.DeviceStatus != "active" {
							t.Fatalf("expected active status, got %q", device.DeviceStatus)
						}
						if device.LastSeen.IsZero() {
							t.Fatalf("expected LastSeen to be set")
						}
						return nil
					})
			},
		},
		{
			name:    "get by id error stops registration",
			topic:   "register/device-4",
			payload: []byte(`{"name":"Camera 4"}`),
			setupMocks: func(mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device-4").
					Return(nil, errors.New("db unavailable"))
			},
		},
		{
			name:    "save error is handled without panic",
			topic:   "register/device-5",
			payload: []byte(`{"name":"Camera 5"}`),
			setupMocks: func(mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device-5").
					Return(nil, mongo.ErrNoDocuments)
				mockDevice.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(errors.New("insert failed"))
			},
		},
		{
			name:    "update error is handled without panic",
			topic:   "register/device-6",
			payload: []byte(`{"name":"Camera 6"}`),
			setupMocks: func(mockDevice *mock_domain.MockDeviceRepository) {
				mockDevice.EXPECT().
					GetByID(gomock.Any(), "device-6").
					Return(&domain.Device{DeviceID: "device-6"}, nil)
				mockDevice.EXPECT().
					Update(gomock.Any(), "device-6", gomock.Any()).
					Return(errors.New("update failed"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPhoto := mock_domain.NewMockPhotoRepository(ctrl)
			mockDevice := mock_domain.NewMockDeviceRepository(ctrl)
			tt.setupMocks(mockDevice)

			b := broker.NewBrokerHandlerWithRepos(mockPhoto, mockDevice)
			msg := &mockMessage{
				topic:   tt.topic,
				payload: tt.payload,
			}

			b.RegisterDevice(nil, msg)
		})
	}
}
