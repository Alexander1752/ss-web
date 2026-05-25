package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/repository"
)

// DeviceController handles HTTP requests for device operations.
// Business logic is delegated to DeviceService.
type DeviceController struct {
	Service *DeviceService
}

func InitDeviceRoutes(db *mongo.Database, mqttClient mqtt.Client, mux *http.ServeMux) {
	deviceController := &DeviceController{
		Service: NewDeviceService(repository.NewDeviceRepository(db), mqttClient),
	}

	mux.Handle("/devices", withAuth(http.HandlerFunc(deviceController.GetDevices)))
	mux.Handle("/devices/switch", withAuth(http.HandlerFunc(deviceController.SwitchDeviceMode)))
	mux.Handle("/devices/command", withAuth(http.HandlerFunc(deviceController.SendCommand)))
}

func (ctlr DeviceController) GetDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role, ok := ctx.Value("role").(string)
	if !ok || role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	devices, err := ctlr.Service.ListDevices(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func (ctlr DeviceController) SwitchDeviceMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role, ok := ctx.Value("role").(string)
	if !ok || role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID   string `json:"id"`
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ctlr.Service.SwitchMode(req.ID, req.Mode); err != nil {
		if errors.Is(err, ErrInvalidDeviceID) || errors.Is(err, ErrInvalidMode) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ctlr DeviceController) SendCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceID string `json:"device_id"`
		Command  string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ctlr.Service.SendCommand(req.DeviceID, req.Command); err != nil {
		if errors.Is(err, ErrInvalidCommand) {
			http.Error(w, "Invalid command. Must be CAPTURE, START-LIVE, or STOP-LIVE", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to publish command", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Command %s sent to device %s", req.Command, req.DeviceID),
	})
}

