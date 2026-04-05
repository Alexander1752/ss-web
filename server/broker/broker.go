package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/otiai10/gosseract/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
	"mqtt-streaming-server/utils"
)

type BrokerHandler struct {
	photoRepository  domain.PhotoRepository
	deviceRepository domain.DeviceRepository
	ocrClient        *gosseract.Client
}

func NewBrokerHandler(db *mongo.Database, ocrClient *gosseract.Client) BrokerHandler {
	return BrokerHandler{
		photoRepository:  repository.NewPhotoRepository(db),
		deviceRepository: repository.NewDeviceRepository(db),
		ocrClient:        ocrClient,
	}
}

func (b BrokerHandler) HandlePhoto(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	var deviceID string
	// topic is ssproject/images/device_id or just ssproject/images
	if topic == "ssproject/images" {
		deviceID = "camera_stream"
	} else if len(topic) > len("ssproject/images/") {
		deviceID = topic[len("ssproject/images/"):]
	} else {
		deviceID = "unknown"
	}

	ctx := context.Background()
	fmt.Println("Received message on topic:", msg.Topic())

	// get registered device
	device, err := b.deviceRepository.GetByID(ctx, deviceID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("Device ID not found: %s. Auto-registering...\n", deviceID)
			// Auto-register the device
			newDevice := &domain.Device{
				DeviceID:     deviceID,
				DeviceName:   "Unknown Device (" + deviceID + ")",
				DeviceStatus: "active",
			}
			if err := b.deviceRepository.Save(ctx, newDevice); err != nil {
				fmt.Printf("Failed to auto-register device: %v\n", err)
				return
			}
			device = newDevice
		} else {
			fmt.Printf("Failed to check device ID: %v\n", err)
			return
		}
	}
	fmt.Printf("Received photo from device: %s\n", device.DeviceName)
	body := msg.Payload()
	_, imageType, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		return
	}
	fmt.Printf("Image type: %s\n", imageType)

	// Extract text from image
	text, err := b.extractTextFromImage(body)
	if err != nil {
		fmt.Printf("Failed to extract text from image: %v\n", err)
		text = "OCR failed"
	}

	// Try to extract structured medical data
	var medicalData *utils.MedicalData
	if utils.IsMedicalCertificate(text) {
		medicalData = utils.ParseMedicalCertificate(text)
		if medicalData != nil {
			fmt.Printf("Extracted medical data: %+v\n", medicalData)
		}
	}

	// UTC timestamp
	timestamp := time.Now().UTC()

	// Create photo with flattened medical data
	photo := &domain.Photo{
		ImageType: imageType,
		Timestamp: timestamp,
		DeviceID:  deviceID,
		Text:      text,
	}

	// Copy medical data fields directly to photo (flattened)
	if medicalData != nil {
		photo.UnitateMedicala = medicalData.UnitateMedicala
		photo.AdresaUnitateMedicala = medicalData.AdresaUnitateMedicala
		photo.TelefonUnitateMedicala = medicalData.TelefonUnitateMedicala
		photo.NumarFisa = medicalData.NumarFisa
		photo.SocietateUnitate = medicalData.SocietateUnitate
		photo.AdresaAngajator = medicalData.AdresaAngajator
		photo.TelefonAngajator = medicalData.TelefonAngajator
		photo.Nume = medicalData.Nume
		photo.Prenume = medicalData.Prenume
		photo.CNP = medicalData.CNP
		photo.ProfesieFunctie = medicalData.ProfesieFunctie
		photo.LocDeMunca = medicalData.LocDeMunca
		photo.TipControl = medicalData.TipControl
		photo.ControlAngajare = medicalData.ControlAngajare
		photo.ControlPeriodic = medicalData.ControlPeriodic
		photo.ControlAdaptare = medicalData.ControlAdaptare
		photo.ControlReluare = medicalData.ControlReluare
		photo.ControlSupraveghere = medicalData.ControlSupraveghere
		photo.ControlAlte = medicalData.ControlAlte

		photo.AvizMedical = medicalData.AvizMedical
		photo.AvizApt = medicalData.AvizApt
		photo.AvizAptConditionat = medicalData.AvizAptConditionat
		photo.AvizInaptTemporar = medicalData.AvizInaptTemporar
		photo.AvizInapt = medicalData.AvizInapt

		photo.Recomandari = medicalData.Recomandari
		photo.Data = medicalData.Data
		photo.DataUrmExaminari = medicalData.DataUrmExaminari
	}

	err = b.photoRepository.Save(ctx, photo)
	if err != nil {
		fmt.Printf("Failed to insert photo into MongoDB: %v\n", err)
		return
	}
	// Save photo object to MinIO
	keyName := fmt.Sprintf("photos/%d.%s", timestamp.Unix(), imageType)
	if err := utils.SaveToMinIO(body, keyName); err != nil {
		fmt.Printf("Failed to save photo to MinIO: %v\n", err)
		return
	}
	fmt.Printf("Photo saved to MinIO with key: %s\n", keyName)
}

func (b BrokerHandler) RegisterDevice(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	// topic is register/device_id
	deviceID := topic[len("register/"):]
	ctx := context.Background()
	fmt.Println("Received message on topic:", msg.Topic())
	body := msg.Payload()
	fmt.Printf("Received device registration: %s\n", body)

	// Parse JSON payload: {"name": "...", "ip": "...", "port": "..."}
	var deviceName, ipAddress, port string
	var registration struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	if err := json.Unmarshal(body, &registration); err == nil && registration.Name != "" {
		deviceName = registration.Name
		ipAddress = registration.IP
		port = registration.Port
	} else {
		deviceName = string(body)
	}

	// Check if device ID already exists
	_, err := b.deviceRepository.GetByID(ctx, deviceID)
	if err != nil && err != mongo.ErrNoDocuments {
		fmt.Printf("Failed to check device ID: %v\n", err)
		return
	}
	if err == mongo.ErrNoDocuments {
		// Device ID does not exist, insert it
		err = b.deviceRepository.Save(ctx, &domain.Device{
			DeviceID:     deviceID,
			DeviceName:   deviceName,
			DeviceStatus: "active",
			IPAddress:    ipAddress,
			Port:         port,
			LastSeen:     time.Now().UTC(),
		})
		if err != nil {
			fmt.Printf("Failed to insert device ID: %v\n", err)
			return
		}
		fmt.Printf("Device registered: %s (IP: %s, Port: %s)\n", deviceID, ipAddress, port)
		return
	}
	// Device ID already exists, update it
	err = b.deviceRepository.Update(ctx, deviceID, &domain.Device{
		DeviceID:     deviceID,
		DeviceName:   deviceName,
		DeviceStatus: "active",
		IPAddress:    ipAddress,
		Port:         port,
		LastSeen:     time.Now().UTC(),
	})
	if err != nil {
		fmt.Printf("Failed to update device ID: %v\n", err)
		return
	}
	fmt.Printf("Device updated: %s (IP: %s, Port: %s)\n", deviceID, ipAddress, port)
}

func (b BrokerHandler) DisconnectDevice(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	// topic is device/id/device_id
	var deviceID string
	if len(topic) > len("device/id/") {
		deviceID = topic[len("device/id/"):]
	} else {
		return
	}

	ctx := context.Background()
	fmt.Println("Received message on topic:", msg.Topic())
	message := string(msg.Payload())
	fmt.Printf("Received device disconnection: %s\n", message)

	if message != "Device Disconnected" {
		fmt.Printf("Invalid disconnection message: %s\n", message)
		return
	}

	device, err := b.deviceRepository.GetByID(ctx, deviceID)
	if err != nil {
		// handle error
		return
	}
	if device.DeviceStatus != "active" {
		return
	}
	err = b.deviceRepository.Update(ctx, deviceID, &domain.Device{
		DeviceID:     deviceID,
		DeviceStatus: "inactive",
		DeviceName:   device.DeviceName,
	})
}

func (b BrokerHandler) extractTextFromImage(imageData []byte) (string, error) {
	// Use the OCR client to extract text from the image
	b.ocrClient.SetImageFromBytes(imageData)
	text, err := b.ocrClient.Text()
	if err != nil {
		return "", fmt.Errorf("failed to extract text from image: %v", err)
	}
	return text, nil
}

func (b BrokerHandler) HandleCommand(_ mqtt.Client, msg mqtt.Message) {
	fmt.Println("Received command on topic:", msg.Topic())
	body := string(msg.Payload())
	fmt.Printf("Command payload: %s\n", body)
}
