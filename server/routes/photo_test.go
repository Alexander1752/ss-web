package routes_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"

	"mqtt-streaming-server/domain"
	mock_domain "mqtt-streaming-server/mocks"
	"mqtt-streaming-server/routes"
)

func TestPhotoController_GetPhotos(t *testing.T) {
	t.Run("no photos and defaults", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
		ctlr := routes.PhotoController{Service: &routes.PhotoService{Repo: mockRepo}}

		var capturedFilters map[string]any
		mockRepo.EXPECT().
			GetPhotos(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, filters map[string]any) ([]*domain.Photo, error) {
				capturedFilters = filters
				return []*domain.Photo{}, nil
			})

		req := httptest.NewRequest(http.MethodGet, "/photos", nil)
		ctx := context.WithValue(req.Context(), "email", "empty@example.com")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		ctlr.GetPhotos(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("expected Content-Type application/json, got %q", ct)
		}

		if capturedFilters == nil {
			t.Fatalf("expected filters to be passed to repository")
		}
		ts, ok := capturedFilters["timestamp"].(map[string]any)
		if !ok {
			t.Fatalf("expected timestamp filter to be a map, got %T", capturedFilters["timestamp"])
		}
		if _, ok := ts["$gte"].(time.Time); !ok {
			t.Fatalf("expected $gte to be time.Time, got %T", ts["$gte"])
		}
		if _, ok := ts["$lte"].(time.Time); !ok {
			t.Fatalf("expected $lte to be time.Time, got %T", ts["$lte"])
		}
	})

	t.Run("success returns photos and tolerates missing storage config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
		ctlr := routes.PhotoController{Service: routes.NewPhotoService(mockRepo)}

		t.Setenv("MINIO_ENDPOINT", "")

		photo := &domain.Photo{
			Timestamp: time.Unix(1600000000, 0),
			ImageType: "png",
		}

		mockRepo.EXPECT().GetPhotos(gomock.Any(), gomock.Any()).Return([]*domain.Photo{photo}, nil)

		req := httptest.NewRequest(http.MethodGet, "/photos", nil)
		rr := httptest.NewRecorder()

		ctlr.GetPhotos(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, `"presigned_url":""`) {
			t.Fatalf("expected empty presigned url when storage is not configured, got %q", body)
		}
	})

	t.Run("filters include text and device_id and use provided start/end", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
		ctlr := routes.PhotoController{Service: routes.NewPhotoService(mockRepo)}

		var capturedFilters map[string]any
		mockRepo.EXPECT().GetPhotos(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, filters map[string]any) ([]*domain.Photo, error) {
				capturedFilters = filters
				return []*domain.Photo{}, nil
			})

		start := strconvFormatInt(time.Now().Add(-2 * time.Hour).UTC().Unix())
		end := strconvFormatInt(time.Now().UTC().Unix())
		url := fmt.Sprintf("/photos?start=%s&end=%s&text=abc&device_id=device-1", start, end)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		ctlr.GetPhotos(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		if capturedFilters == nil {
			t.Fatalf("expected filters to be captured")
		}
		if _, ok := capturedFilters["text"].(map[string]any); !ok {
			t.Fatalf("expected text filter to be a map, got %T", capturedFilters["text"])
		}
		if did, ok := capturedFilters["device_id"].(string); !ok || did != "device-1" {
			t.Fatalf("expected device_id filter to equal 'device-1', got %v", capturedFilters["device_id"])
		}
	})
}

func strconvFormatInt(i int64) string { return fmt.Sprintf("%d", i) }

func TestPhotoController_GetPhotos_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{Service: &routes.PhotoService{Repo: mockRepo}}

	req := httptest.NewRequest(http.MethodPost, "/photos", nil)
	rr := httptest.NewRecorder()

	ctlr.GetPhotos(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestPhotoController_GetPhotos_InvalidTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{Service: &routes.PhotoService{Repo: mockRepo}}

	req := httptest.NewRequest(http.MethodGet, "/photos?start=invalid&end=invalid", nil)
	rr := httptest.NewRecorder()

	ctlr.GetPhotos(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid start timestamp") && !strings.Contains(rr.Body.String(), "Invalid end timestamp") {
		t.Errorf("expected body to contain 'Invalid start timestamp' or 'Invalid end timestamp', got %q", rr.Body.String())
	}
}

func TestPhotoController_DeletePhoto_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{Service: routes.NewPhotoService(mockRepo)}

	req := httptest.NewRequest(http.MethodGet, "/photos/123", nil)
	rr := httptest.NewRecorder()

	ctlr.DeletePhoto(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestPhotoController_DeletePhoto_BadRequestsAndErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{Service: routes.NewPhotoService(mockRepo)}

	req := httptest.NewRequest(http.MethodDelete, "/photos/", nil)
	rr := httptest.NewRecorder()
	ctlr.DeletePhoto(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for missing id, got %d", http.StatusBadRequest, rr.Code)
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), "notfound").Return(nil, mongo.ErrNoDocuments)
	req = httptest.NewRequest(http.MethodDelete, "/photos/notfound", nil)
	rr = httptest.NewRecorder()
	ctlr.DeletePhoto(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d for not found, got %d", http.StatusNotFound, rr.Code)
	}

	photo := &domain.Photo{Timestamp: time.Unix(1600000001, 0), ImageType: "jpg"}
	mockRepo.EXPECT().GetByID(gomock.Any(), "did").Return(photo, nil)
	mockRepo.EXPECT().Delete(gomock.Any(), "did").Return(errors.New("db delete error"))
	req = httptest.NewRequest(http.MethodDelete, "/photos/did", nil)
	rr = httptest.NewRecorder()
	ctlr.DeletePhoto(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for delete failure, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestPhotoController_DeletePhoto_SuccessToleratesStorageDeleteFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{Service: routes.NewPhotoService(mockRepo)}

	timestamp := int64(1600000002)
	imageType := "png"
	t.Setenv("MINIO_ENDPOINT", "")

	photo := &domain.Photo{Timestamp: time.Unix(timestamp, 0), ImageType: imageType}
	mockRepo.EXPECT().GetByID(gomock.Any(), "goodid").Return(photo, nil)
	mockRepo.EXPECT().Delete(gomock.Any(), "goodid").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/photos/goodid", nil)
	rr := httptest.NewRecorder()

	ctlr.DeletePhoto(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Photo deleted successfully") {
		t.Fatalf("expected success message, got %q", rr.Body.String())
	}
}

func TestPhotoController_DeleteAllPhotos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{Service: routes.NewPhotoService(mockRepo)}

	req := httptest.NewRequest(http.MethodGet, "/photos/all", nil)
	rr := httptest.NewRecorder()
	ctlr.DeleteAllPhotos(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d for method not allowed, got %d", http.StatusMethodNotAllowed, rr.Code)
	}

	mockRepo.EXPECT().DeleteAll(gomock.Any()).Return(int64(0), errors.New("db error"))
	req = httptest.NewRequest(http.MethodDelete, "/photos/all", nil)
	rr = httptest.NewRecorder()
	ctlr.DeleteAllPhotos(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for delete all failure, got %d", http.StatusInternalServerError, rr.Code)
	}

	t.Setenv("MINIO_ENDPOINT", "")

	mockRepo.EXPECT().DeleteAll(gomock.Any()).Return(int64(3), nil)
	req = httptest.NewRequest(http.MethodDelete, "/photos/all", nil)
	rr = httptest.NewRecorder()
	ctlr.DeleteAllPhotos(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d for successful delete all, got %d", http.StatusOK, rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response json: %v", err)
	}
	if v, ok := resp["deleted"].(float64); !ok || int64(v) != 3 {
		t.Fatalf("expected deleted count 3, got %v", resp["deleted"])
	}
}
