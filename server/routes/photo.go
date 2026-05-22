package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/repository"
)

// PhotoController handles HTTP requests for photo operations.
// Business logic is delegated to PhotoService.
type PhotoController struct {
	Service *PhotoService
}

func InitPhotoRoutes(db *mongo.Database, mux *http.ServeMux) {
	photoController := &PhotoController{
		Service: NewPhotoService(repository.NewPhotoRepository(db)),
	}

	mux.Handle("/photos", withAuth(http.HandlerFunc(photoController.GetPhotos)))
	mux.Handle("/photos/all", withAuth(http.HandlerFunc(photoController.DeleteAllPhotos)))
	mux.Handle("/photos/", withAuth(http.HandlerFunc(photoController.DeletePhoto)))
}

func (ctlr PhotoController) GetPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	text := r.URL.Query().Get("text")
	deviceID := r.URL.Query().Get("device_id")

	if start == "" {
		start = strconv.FormatInt(time.Now().Add(-24*time.Hour).UTC().Unix(), 10)
	}
	if end == "" {
		end = strconv.FormatInt(time.Now().UTC().Unix(), 10)
	}

	startInt, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		http.Error(w, "Invalid start timestamp "+err.Error(), http.StatusBadRequest)
		return
	}

	endInt, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		http.Error(w, "Invalid end timestamp "+err.Error(), http.StatusBadRequest)
		return
	}

	photos, err := ctlr.Service.ListPhotos(r.Context(), startInt, endInt, text, deviceID)
	if err != nil {
		http.Error(w, "Failed to fetch photos: ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func (ctlr PhotoController) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/photos/")
	if path == "" {
		http.Error(w, "Photo ID required", http.StatusBadRequest)
		return
	}

	if err := ctlr.Service.DeletePhoto(r.Context(), path); err != nil {
		if errors.Is(err, ErrPhotoNotFound) {
			http.Error(w, "Photo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Photo deleted successfully"})
}

func (ctlr PhotoController) DeleteAllPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := ctlr.Service.DeleteAllPhotos(r.Context())
	if err != nil {
		http.Error(w, "Failed to delete photos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "All photos deleted successfully",
		"deleted": count,
	})
}

