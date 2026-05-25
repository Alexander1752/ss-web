package routes

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"mqtt-streaming-server/domain"
)

// ErrInvalidCommand is returned when an unrecognised camera command is requested.
var ErrInvalidCommand = errors.New("invalid command")

// ErrInvalidDeviceID is returned when a device ID contains characters that are
// unsafe in MQTT topics (/, +, # or anything outside [a-zA-Z0-9_-]).
var ErrInvalidDeviceID = errors.New("invalid device ID")

// ErrInvalidMode is returned when the requested mode is not in the allowed list.
var ErrInvalidMode = errors.New("invalid mode")

var validCommands = map[string]bool{
	"CAPTURE":    true,
	"START-LIVE": true,
	"STOP-LIVE":  true,
}

var validModes = map[string]bool{
	"LIVE":   true,
	"NORMAL": true,
}

// deviceIDRegex allows only alphanumeric characters, hyphens, and underscores,
// preventing injection of MQTT wildcard characters (+, #) and topic separators (/).
var deviceIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

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
// Returns ErrInvalidDeviceID if deviceID contains unsafe characters, ErrInvalidMode if
// mode is not one of the allowed values.
func (s *DeviceService) SwitchMode(deviceID, mode string) error {
	if !deviceIDRegex.MatchString(deviceID) {
		return fmt.Errorf("%w: %q", ErrInvalidDeviceID, deviceID)
	}
	if !validModes[mode] {
		return fmt.Errorf("%w: %q", ErrInvalidMode, mode)
	}
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
