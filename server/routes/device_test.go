package routes

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/mock/gomock"

	"mqtt-streaming-server/domain"
	mock_domain "mqtt-streaming-server/mocks"
)

type mockMQTTToken struct {
	err error
}

func (t *mockMQTTToken) Wait() bool {
	return true
}

func (t *mockMQTTToken) WaitTimeout(_ time.Duration) bool {
	return true
}

func (t *mockMQTTToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (t *mockMQTTToken) Error() error {
	return t.err
}

type mockMQTTClient struct {
	publishErr    error
	lastTopic     string
	lastQoS       byte
	lastRetained  bool
	lastPayload   any
	publishCalled bool
}

func (c *mockMQTTClient) IsConnected() bool {
	return true
}

func (c *mockMQTTClient) IsConnectionOpen() bool {
	return true
}

func (c *mockMQTTClient) Connect() mqtt.Token {
	return &mockMQTTToken{}
}

func (c *mockMQTTClient) Disconnect(_ uint) {}

func (c *mockMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	c.publishCalled = true
	c.lastTopic = topic
	c.lastQoS = qos
	c.lastRetained = retained
	c.lastPayload = payload
	return &mockMQTTToken{err: c.publishErr}
}

func (c *mockMQTTClient) Subscribe(_ string, _ byte, _ mqtt.MessageHandler) mqtt.Token {
	return &mockMQTTToken{}
}

func (c *mockMQTTClient) SubscribeMultiple(_ map[string]byte, _ mqtt.MessageHandler) mqtt.Token {
	return &mockMQTTToken{}
}

func (c *mockMQTTClient) Unsubscribe(_ ...string) mqtt.Token {
	return &mockMQTTToken{}
}

func (c *mockMQTTClient) AddRoute(_ string, _ mqtt.MessageHandler) {}

func (c *mockMQTTClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}

func TestDeviceController_GetDevices(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		ctlr := DeviceController{}

		req := httptest.NewRequest(http.MethodPost, "/devices", nil)
		rr := httptest.NewRecorder()

		ctlr.GetDevices(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Method not allowed") {
			t.Fatalf("expected body to contain method error, got %q", rr.Body.String())
		}
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockDeviceRepository(ctrl)
		ctlr := DeviceController{DeviceRepository: mockRepo}

		mockRepo.EXPECT().GetAllDevices(gomock.Any()).Return(nil, errors.New("db error"))

		req := httptest.NewRequest(http.MethodGet, "/devices", nil)
		rr := httptest.NewRecorder()

		ctlr.GetDevices(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Failed to fetch devices") {
			t.Fatalf("expected body to contain repository error, got %q", rr.Body.String())
		}
	})

	t.Run("success returns json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockDeviceRepository(ctrl)
		ctlr := DeviceController{DeviceRepository: mockRepo}

		mockRepo.EXPECT().GetAllDevices(gomock.Any()).Return([]*domain.Device{
			{ID: "dev-1", DeviceName: "iPhone"},
		}, nil)

		req := httptest.NewRequest(http.MethodGet, "/devices", nil)
		rr := httptest.NewRecorder()

		ctlr.GetDevices(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("expected JSON content type, got %q", got)
		}
		if !strings.Contains(rr.Body.String(), "iPhone") {
			t.Fatalf("expected response to contain device payload, got %q", rr.Body.String())
		}
	})
}

func TestDeviceController_SwitchDeviceMode(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		ctlr := DeviceController{}

		req := httptest.NewRequest(http.MethodGet, "/devices/switch", nil)
		rr := httptest.NewRecorder()

		ctlr.SwitchDeviceMode(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Method not allowed") {
			t.Fatalf("expected method error body, got %q", rr.Body.String())
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		ctlr := DeviceController{mqttClient: &mockMQTTClient{}}

		req := httptest.NewRequest(http.MethodPost, "/devices/switch", strings.NewReader("invalid json"))
		rr := httptest.NewRecorder()

		ctlr.SwitchDeviceMode(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Invalid request body") {
			t.Fatalf("expected invalid body message, got %q", rr.Body.String())
		}
	})

	t.Run("mqtt publish error", func(t *testing.T) {
		mqttClient := &mockMQTTClient{publishErr: errors.New("publish failed")}
		ctlr := DeviceController{mqttClient: mqttClient}

		req := httptest.NewRequest(http.MethodPost, "/devices/switch", strings.NewReader(`{"id":"dev-1","mode":"LIVE"}`))
		rr := httptest.NewRecorder()

		ctlr.SwitchDeviceMode(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Failed to publish message") {
			t.Fatalf("expected mqtt error message, got %q", rr.Body.String())
		}
		if mqttClient.lastTopic != "setup/dev-1" {
			t.Fatalf("expected topic setup/dev-1, got %q", mqttClient.lastTopic)
		}
		if mqttClient.lastPayload != "start LIVE" {
			t.Fatalf("expected payload 'start LIVE', got %v", mqttClient.lastPayload)
		}
	})

	t.Run("success", func(t *testing.T) {
		mqttClient := &mockMQTTClient{}
		ctlr := DeviceController{mqttClient: mqttClient}

		req := httptest.NewRequest(http.MethodPost, "/devices/switch", strings.NewReader(`{"id":"dev-2","mode":"NORMAL"}`))
		rr := httptest.NewRecorder()

		ctlr.SwitchDeviceMode(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		if !mqttClient.publishCalled {
			t.Fatalf("expected publish to be called")
		}
		if mqttClient.lastTopic != "setup/dev-2" {
			t.Fatalf("expected topic setup/dev-2, got %q", mqttClient.lastTopic)
		}
		if mqttClient.lastPayload != "start NORMAL" {
			t.Fatalf("expected payload 'start NORMAL', got %v", mqttClient.lastPayload)
		}
	})
}

func TestDeviceController_SendCommand(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		ctlr := DeviceController{}

		req := httptest.NewRequest(http.MethodGet, "/devices/command", nil)
		rr := httptest.NewRecorder()

		ctlr.SendCommand(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Method not allowed") {
			t.Fatalf("expected method error body, got %q", rr.Body.String())
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		ctlr := DeviceController{mqttClient: &mockMQTTClient{}}

		req := httptest.NewRequest(http.MethodPost, "/devices/command", strings.NewReader("invalid json"))
		rr := httptest.NewRecorder()

		ctlr.SendCommand(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Invalid request body") {
			t.Fatalf("expected invalid body message, got %q", rr.Body.String())
		}
	})

	t.Run("invalid command", func(t *testing.T) {
		ctlr := DeviceController{mqttClient: &mockMQTTClient{}}

		req := httptest.NewRequest(http.MethodPost, "/devices/command", strings.NewReader(`{"device_id":"dev-1","command":"REBOOT"}`))
		rr := httptest.NewRecorder()

		ctlr.SendCommand(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Invalid command") {
			t.Fatalf("expected invalid command message, got %q", rr.Body.String())
		}
	})

	t.Run("mqtt publish error", func(t *testing.T) {
		mqttClient := &mockMQTTClient{publishErr: errors.New("publish failed")}
		ctlr := DeviceController{mqttClient: mqttClient}

		req := httptest.NewRequest(http.MethodPost, "/devices/command", strings.NewReader(`{"device_id":"dev-1","command":"CAPTURE"}`))
		rr := httptest.NewRecorder()

		ctlr.SendCommand(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Failed to publish command") {
			t.Fatalf("expected publish error message, got %q", rr.Body.String())
		}
		if mqttClient.lastTopic != "ssproject/commands" {
			t.Fatalf("expected topic ssproject/commands, got %q", mqttClient.lastTopic)
		}
		if mqttClient.lastPayload != "CAPTURE" {
			t.Fatalf("expected payload CAPTURE, got %v", mqttClient.lastPayload)
		}
	})

	t.Run("success returns json", func(t *testing.T) {
		mqttClient := &mockMQTTClient{}
		ctlr := DeviceController{mqttClient: mqttClient}

		req := httptest.NewRequest(http.MethodPost, "/devices/command", strings.NewReader(`{"device_id":"dev-42","command":"START-LIVE"}`))
		rr := httptest.NewRecorder()

		ctlr.SendCommand(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("expected JSON content type, got %q", got)
		}
		if !strings.Contains(rr.Body.String(), `"status":"success"`) {
			t.Fatalf("expected success JSON status, got %q", rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "Command START-LIVE sent to device dev-42") {
			t.Fatalf("expected success message, got %q", rr.Body.String())
		}
		if mqttClient.lastTopic != "ssproject/commands" {
			t.Fatalf("expected topic ssproject/commands, got %q", mqttClient.lastTopic)
		}
		if mqttClient.lastPayload != "START-LIVE" {
			t.Fatalf("expected payload START-LIVE, got %v", mqttClient.lastPayload)
		}
	})
}