package broker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
	"mqtt-streaming-server/utils"
)

type BrokerHandler struct {
	photoRepository  domain.PhotoRepository
	deviceRepository domain.DeviceRepository
}

func NewBrokerHandler(db *mongo.Database) BrokerHandler {
	return BrokerHandler{
		photoRepository:  repository.NewPhotoRepository(db),
		deviceRepository: repository.NewDeviceRepository(db),
	}
}

func NewBrokerHandlerWithRepos(photoRepo domain.PhotoRepository, deviceRepo domain.DeviceRepository) BrokerHandler {
	return BrokerHandler{
		photoRepository:  photoRepo,
		deviceRepository: deviceRepo,
	}
}

func (b BrokerHandler) HandlePhoto(client mqtt.Client, msg mqtt.Message) {
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

	// UTC timestamp
	timestamp := time.Now().UTC()

	// Initial placeholder for Text before OCR completes
	text := "OCR processing..."

	// Create photo with basic data
	photo := &domain.Photo{
		ImageType: imageType,
		Timestamp: timestamp,
		DeviceID:  deviceID,
		Text:      text,
	}

	err = b.photoRepository.Save(ctx, photo)
	if err != nil {
		fmt.Printf("Failed to insert photo into MongoDB: %v\n", err)
		return
	}

	// Publish to OCR service
	ocrReq := map[string]string{
		"photo_id":     photo.ID.Hex(),
		"image_base64": base64.StdEncoding.EncodeToString(body),
	}
	reqBody, _ := json.Marshal(ocrReq)
	token := client.Publish("ssproject/ocr/requests", 0, false, reqBody)
	if !token.WaitTimeout(5 * time.Second) {
		fmt.Printf("Timed out publishing OCR request for photo %s\n", photo.ID.Hex())
		return
	}
	if err := token.Error(); err != nil {
		fmt.Printf("Failed to publish OCR request for photo %s: %v\n", photo.ID.Hex(), err)
		return
	}
	fmt.Printf("Sent OCR request for photo %s\n", photo.ID.Hex())

	// Save photo object to MinIO
	keyName := fmt.Sprintf("photos/%d.%s", timestamp.Unix(), imageType)
	if err := utils.SaveToMinIO(body, keyName); err != nil {
		fmt.Printf("Failed to save photo to MinIO: %v\n", err)
		return
	}
	fmt.Printf("Photo saved to MinIO with key: %s\n", keyName)
}

func (b BrokerHandler) HandleOCRResult(_ mqtt.Client, msg mqtt.Message) {
	fmt.Println("Received OCR result on topic:", msg.Topic())
	var result struct {
		PhotoID string `json:"photo_id"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal(msg.Payload(), &result); err != nil {
		fmt.Printf("Failed to unmarshal OCR result: %v\n", err)
		return
	}

	ctx := context.Background()
	objID, err := primitive.ObjectIDFromHex(result.PhotoID)
	if err != nil {
		fmt.Printf("Invalid PhotoID %s: %v\n", result.PhotoID, err)
		return
	}

	updates := map[string]any{
		"text": result.Text,
	}

	if utils.IsMedicalCertificate(result.Text) {
		medicalData := utils.ParseMedicalCertificate(result.Text)
		if medicalData != nil {
			fmt.Printf("Extracted medical data for %s: %+v\n", result.PhotoID, medicalData)
			updates["unitate_medicala"] = medicalData.UnitateMedicala
			updates["adresa_unitate_medicala"] = medicalData.AdresaUnitateMedicala
			updates["telefon_unitate_medicala"] = medicalData.TelefonUnitateMedicala
			updates["numar_fisa"] = medicalData.NumarFisa
			updates["societate_unitate"] = medicalData.SocietateUnitate
			updates["adresa_angajator"] = medicalData.AdresaAngajator
			updates["telefon_angajator"] = medicalData.TelefonAngajator
			updates["nume"] = medicalData.Nume
			updates["prenume"] = medicalData.Prenume
			updates["cnp"] = medicalData.CNP
			updates["profesie_functie"] = medicalData.ProfesieFunctie
			updates["loc_de_munca"] = medicalData.LocDeMunca
			updates["tip_control"] = medicalData.TipControl
			updates["control_angajare"] = medicalData.ControlAngajare
			updates["control_periodic"] = medicalData.ControlPeriodic
			updates["control_adaptare"] = medicalData.ControlAdaptare
			updates["control_reluare"] = medicalData.ControlReluare
			updates["control_supraveghere"] = medicalData.ControlSupraveghere
			updates["control_alte"] = medicalData.ControlAlte

			updates["aviz_medical"] = medicalData.AvizMedical
			updates["aviz_apt"] = medicalData.AvizApt
			updates["aviz_apt_conditionat"] = medicalData.AvizAptConditionat
			updates["aviz_inapt_temporar"] = medicalData.AvizInaptTemporar
			updates["aviz_inapt"] = medicalData.AvizInapt

			updates["recomandari"] = medicalData.Recomandari
			updates["data"] = medicalData.Data
			updates["data_urm_examinari"] = medicalData.DataUrmExaminari
		}
	}

	if err := b.photoRepository.UpdatePhotoTextAndMedicalData(ctx, objID, updates); err != nil {
		fmt.Printf("Failed to update photo OCR result: %v\n", err)
	} else {
		fmt.Printf("Successfully updated photo %s with OCR data\n", result.PhotoID)
	}
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

func (b BrokerHandler) HandleCommand(_ mqtt.Client, msg mqtt.Message) {
	fmt.Println("Received command on topic:", msg.Topic())
	body := string(msg.Payload())
	fmt.Printf("Command payload: %s\n", body)
}
