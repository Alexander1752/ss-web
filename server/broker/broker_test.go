package broker

import (
	"context"
	"errors"
	"testing"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
)

type testMQTTMessage struct {
	topic   string
	payload []byte
}

func (m testMQTTMessage) Duplicate() bool { return false }
func (m testMQTTMessage) Qos() byte       { return 0 }
func (m testMQTTMessage) Retained() bool  { return false }
func (m testMQTTMessage) Topic() string   { return m.topic }
func (m testMQTTMessage) MessageID() uint16 {
	return 0
}
func (m testMQTTMessage) Payload() []byte { return m.payload }
func (m testMQTTMessage) Ack()            {}

type spyDeviceRepository struct {
	getByIDDevice *domain.Device
	getByIDErr    error
	updateErr     error
	saveErr       error

	lastGetByID string
	lastUpdateID string
	lastUpdate  *domain.Device
	lastSave    *domain.Device

	updateCalls int
	saveCalls   int
}

func (s *spyDeviceRepository) GetAllDevices(context.Context) ([]*domain.Device, error) {
	return nil, nil
}

func (s *spyDeviceRepository) GetByID(_ context.Context, id string) (*domain.Device, error) {
	s.lastGetByID = id
	return s.getByIDDevice, s.getByIDErr
}

func (s *spyDeviceRepository) Update(_ context.Context, id string, device *domain.Device) error {
	s.updateCalls++
	s.lastUpdateID = id
	s.lastUpdate = device
	return s.updateErr
}

func (s *spyDeviceRepository) Save(_ context.Context, device *domain.Device) error {
	s.saveCalls++
	s.lastSave = device
	return s.saveErr
}

type noopPhotoRepository struct{}

func (noopPhotoRepository) GetPhotos(context.Context, map[string]any) ([]*domain.Photo, error) { return nil, nil }
func (noopPhotoRepository) GetByID(context.Context, string) (*domain.Photo, error)             { return nil, nil }
func (noopPhotoRepository) Save(context.Context, *domain.Photo) error                           { return nil }
func (noopPhotoRepository) Delete(context.Context, string) error                                { return nil }
func (noopPhotoRepository) DeleteAll(context.Context) (int64, error)                            { return 0, nil }

func TestBrokerHandler_RegisterDevice_ParsingAndFallback(t *testing.T) {
	t.Run("parses json payload for new device", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{getByIDErr: mongo.ErrNoDocuments}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		msg := testMQTTMessage{topic: "register/dev-1", payload: []byte(`{"name":"Sensor 1","ip":"10.0.0.2","port":"1883"}`)}
		b.RegisterDevice(nil, msg)

		if deviceRepo.saveCalls != 1 {
			t.Fatalf("expected one save call, got %d", deviceRepo.saveCalls)
		}
		if deviceRepo.lastSave == nil {
			t.Fatalf("expected saved device")
		}
		if deviceRepo.lastSave.DeviceID != "dev-1" {
			t.Fatalf("expected device_id dev-1, got %q", deviceRepo.lastSave.DeviceID)
		}
		if deviceRepo.lastSave.DeviceName != "Sensor 1" {
			t.Fatalf("expected parsed name Sensor 1, got %q", deviceRepo.lastSave.DeviceName)
		}
		if deviceRepo.lastSave.IPAddress != "10.0.0.2" || deviceRepo.lastSave.Port != "1883" {
			t.Fatalf("expected parsed ip/port, got ip=%q port=%q", deviceRepo.lastSave.IPAddress, deviceRepo.lastSave.Port)
		}
	})

	t.Run("falls back to raw payload when json invalid", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{getByIDErr: mongo.ErrNoDocuments}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		msg := testMQTTMessage{topic: "register/dev-raw", payload: []byte("Legacy Device Name")}
		b.RegisterDevice(nil, msg)

		if deviceRepo.saveCalls != 1 {
			t.Fatalf("expected one save call, got %d", deviceRepo.saveCalls)
		}
		if deviceRepo.lastSave.DeviceName != "Legacy Device Name" {
			t.Fatalf("expected fallback device name from raw payload, got %q", deviceRepo.lastSave.DeviceName)
		}
	})

	t.Run("updates existing device", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{getByIDDevice: &domain.Device{DeviceID: "dev-2", DeviceStatus: "active"}}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		msg := testMQTTMessage{topic: "register/dev-2", payload: []byte(`{"name":"Updated","ip":"10.0.0.3","port":"1993"}`)}
		b.RegisterDevice(nil, msg)

		if deviceRepo.updateCalls != 1 {
			t.Fatalf("expected one update call, got %d", deviceRepo.updateCalls)
		}
		if deviceRepo.lastUpdateID != "dev-2" {
			t.Fatalf("expected update for dev-2, got %q", deviceRepo.lastUpdateID)
		}
		if deviceRepo.lastUpdate == nil || deviceRepo.lastUpdate.DeviceName != "Updated" {
			t.Fatalf("expected updated device payload, got %+v", deviceRepo.lastUpdate)
		}
	})

	t.Run("stops on repository lookup error", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{getByIDErr: errors.New("db error")}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		msg := testMQTTMessage{topic: "register/dev-err", payload: []byte(`{"name":"X"}`)}
		b.RegisterDevice(nil, msg)

		if deviceRepo.saveCalls != 0 || deviceRepo.updateCalls != 0 {
			t.Fatalf("expected no save/update on lookup error")
		}
	})
}

func TestBrokerHandler_DisconnectDevice_InvalidPayloadAndFallbacks(t *testing.T) {
	t.Run("ignores invalid disconnection payload", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{getByIDDevice: &domain.Device{DeviceID: "dev-1", DeviceStatus: "active", DeviceName: "Cam"}}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		b.DisconnectDevice(nil, testMQTTMessage{topic: "device/id/dev-1", payload: []byte("bye")})

		if deviceRepo.updateCalls != 0 {
			t.Fatalf("expected no update for invalid payload")
		}
	})

	t.Run("ignores short topic", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		b.DisconnectDevice(nil, testMQTTMessage{topic: "device/id/", payload: []byte("Device Disconnected")})

		if deviceRepo.lastGetByID != "" {
			t.Fatalf("expected no lookup for short topic")
		}
	})

	t.Run("marks active device inactive", func(t *testing.T) {
		deviceRepo := &spyDeviceRepository{getByIDDevice: &domain.Device{DeviceID: "dev-3", DeviceStatus: "active", DeviceName: "Cam 3"}}
		b := BrokerHandler{deviceRepository: deviceRepo, photoRepository: noopPhotoRepository{}}

		b.DisconnectDevice(nil, testMQTTMessage{topic: "device/id/dev-3", payload: []byte("Device Disconnected")})

		if deviceRepo.updateCalls != 1 {
			t.Fatalf("expected one update call, got %d", deviceRepo.updateCalls)
		}
		if deviceRepo.lastUpdate == nil || deviceRepo.lastUpdate.DeviceStatus != "inactive" {
			t.Fatalf("expected inactive update payload, got %+v", deviceRepo.lastUpdate)
		}
	})
}
