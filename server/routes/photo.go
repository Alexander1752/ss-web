package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
	"mqtt-streaming-server/utils"
)

type PhotoController struct {
	PhotoRepository domain.PhotoRepository
}

func InitPhotoRoutes(db *mongo.Database, mux *http.ServeMux) {
	photoController := &PhotoController{
		PhotoRepository: repository.NewPhotoRepository(db),
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

	ctx := r.Context()

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

	filters := map[string]any{
		"timestamp": map[string]any{
			"$gte": time.Unix(startInt, 0),
			"$lte": time.Unix(endInt, 0),
		},
	}

	if text != "" {
		filters["text"] = map[string]any{
			"$regex":   text,
			"$options": "i",
		}
	}

	if deviceID != "" {
		filters["device_id"] = deviceID
	}

	photos, err := ctlr.PhotoRepository.GetPhotos(ctx, filters)
	if err != nil {
		fmt.Println("Error fetching photos:", err)
		http.Error(w, "Failed to fetch photos: ", http.StatusInternalServerError)
		return
	}

	for _, photo := range photos {
		keyName := fmt.Sprintf("photos/%d.%s", photo.Timestamp.Unix(), photo.ImageType)
		photo.PresignedURL = utils.GetLocalURL(keyName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func (ctlr PhotoController) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Extract photo ID from URL path: /photos/{id}
	path := strings.TrimPrefix(r.URL.Path, "/photos/")
	if path == "" {
		http.Error(w, "Photo ID required", http.StatusBadRequest)
		return
	}
	photoID := path

	// Get the photo to find the file name
	photo, err := ctlr.PhotoRepository.GetByID(ctx, photoID)
	if err != nil {
		fmt.Println("Error getting photo:", err)
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Delete from database
	err = ctlr.PhotoRepository.Delete(ctx, photoID)
	if err != nil {
		fmt.Println("Error deleting photo:", err)
		http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
		return
	}

	// Delete the image file from local storage
	fileName := fmt.Sprintf("uploads/photos/%d.%s", photo.Timestamp.Unix(), photo.ImageType)
	if err := os.Remove(fileName); err != nil {
		fmt.Printf("Warning: Could not delete file %s: %v\n", fileName, err)
		// Don't fail the request - the DB record is already deleted
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Photo deleted successfully"})
}

func (ctlr PhotoController) DeleteAllPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Delete all photos from database
	deletedCount, err := ctlr.PhotoRepository.DeleteAll(ctx)
	if err != nil {
		fmt.Println("Error deleting all photos:", err)
		http.Error(w, "Failed to delete photos", http.StatusInternalServerError)
		return
	}

	// Delete all image files from uploads/photos directory
	photosDir := "uploads/photos"
	files, err := filepath.Glob(filepath.Join(photosDir, "*"))
	if err == nil {
		for _, f := range files {
			os.Remove(f)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "All photos deleted successfully",
		"deleted": deletedCount,
	})
}
