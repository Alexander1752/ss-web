package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
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
	mux.Handle("/photos/upload", withAuth(http.HandlerFunc(photoController.UploadPhoto)))
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
		fmt.Println("Error fetching photos:", err)
		http.Error(w, "Failed to fetch photos", http.StatusInternalServerError)
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

	if role, _ := r.Context().Value("role").(string); role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

func (ctlr PhotoController) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form with a max memory of 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract device_id (will be empty if not provided)
	deviceID := r.FormValue("device_id")

	// Extract the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the file bytes
	photoBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Extract the extension to use as image type (e.g., "jpg", "png")
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(handler.Filename)), ".")
	if ext == "" {
		ext = "jpg" // Fallback
	}

	// Hand off the pure data to the Service layer
	err = ctlr.Service.UploadPhoto(r.Context(), deviceID, ext, photoBytes)
	if err != nil {
		http.Error(w, "Failed to process upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Photo sent for processing",
	})
}
