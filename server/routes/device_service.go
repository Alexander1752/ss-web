package routes

import (
	"context"
	"errors"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"mqtt-streaming-server/domain"
)

// ErrInvalidCommand is returned when an unrecognised camera command is requested.
var ErrInvalidCommand = errors.New("invalid command")

var validCommands = map[string]bool{
	"CAPTURE":    true,
	"START-LIVE": true,
	"STOP-LIVE":  true,
}

// DeviceService contains the business logic for device operations.
type DeviceService struct {
	Repo       domain.DeviceRepository
	MqttClient mqtt.Client
}

// NewDeviceService creates a new DeviceService.
func NewDeviceService(repo domain.DeviceRepository, mqttClient mqtt.Client) *DeviceService {
	return &DeviceService{Repo: repo, MqttClient: mqttClient}
}

// ListDevices returns all registered devices from the repository.
func (s *DeviceService) ListDevices(ctx context.Context) ([]*domain.Device, error) {
	return s.Repo.GetAllDevices(ctx)
}

// SwitchMode publishes a mode-change command to the device-specific MQTT setup topic.
func (s *DeviceService) SwitchMode(deviceID, mode string) error {
	topic := fmt.Sprintf("setup/%s", deviceID)
	token := s.MqttClient.Publish(topic, 0, false, "start "+mode)
	token.Wait()
	return token.Error()
}

// SendCommand validates and publishes a camera command to the shared commands topic.
// Returns ErrInvalidCommand if the command is not one of CAPTURE, START-LIVE, STOP-LIVE.
func (s *DeviceService) SendCommand(deviceID, command string) error {
	if !validCommands[command] {
		return fmt.Errorf("%w: %s", ErrInvalidCommand, command)
	}
	token := s.MqttClient.Publish("ssproject/commands", 0, false, command)
	token.Wait()
	return token.Error()
}
