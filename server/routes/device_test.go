package routes_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"mqtt-streaming-server/domain"
	mock_domain "mqtt-streaming-server/mocks"
	"mqtt-streaming-server/routes"
)

func TestDeviceController_GetDevices(t *testing.T) {
	tests := []struct {
		name             string
		userEmail        string
		userRole         string
		mockDevices      []*domain.Device
		mockError        error
		expectedStatus   int
		expectedContains string
	}{
		{
			name:      "successful device fetch",
			userEmail: "user@example.com",
			userRole:  "admin",
			mockDevices: []*domain.Device{
				{ID: "dev-1", DeviceName: "iPhone"},
				{ID: "dev-2", DeviceName: "MacBook"},
			},
			expectedStatus:   http.StatusOK,
			expectedContains: "iPhone",
		},
		{
			name:             "no devices",
			userEmail:        "empty@example.com",
			userRole:         "admin",
			mockDevices:      []*domain.Device{},
			expectedStatus:   http.StatusOK,
			expectedContains: "[]",
		},
		{
			name:             "repository error",
			userEmail:        "fail@example.com",
			userRole:         "admin",
			mockError:        errors.New("db error"),
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: "Failed to fetch devices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockDeviceRepository(ctrl)
			ctlr := routes.DeviceController{DeviceRepository: mockRepo}

			req := httptest.NewRequest(http.MethodGet, "/devices", nil)
			ctx := context.WithValue(req.Context(), "email", tt.userEmail)
			ctx = context.WithValue(ctx, "role", tt.userRole)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			if tt.mockDevices != nil || tt.mockError != nil {
				mockRepo.EXPECT().
					GetAllDevices(gomock.Any()).
					Return(tt.mockDevices, tt.mockError)
			}

			ctlr.GetDevices(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedContains, rr.Body.String())
			}
		})
	}
}

func TestDeviceController_SwitchDeviceMode_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockDeviceRepository(ctrl)
	ctlr := routes.DeviceController{DeviceRepository: mockRepo}

	url := "/devices/switch"
	req := httptest.NewRequest(http.MethodGet, url, nil) // Using GET instead of POST
	rr := httptest.NewRecorder()

	ctlr.SwitchDeviceMode(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestDeviceController_GetDevices_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockDeviceRepository(ctrl)
	ctlr := routes.DeviceController{DeviceRepository: mockRepo}

	req := httptest.NewRequest(http.MethodPost, "/devices", nil) // Using POST instead of GET
	rr := httptest.NewRecorder()

	ctlr.GetDevices(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestDeviceController_SwitchDeviceMode_InvalidRequestBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockDeviceRepository(ctrl)
	ctlr := routes.DeviceController{DeviceRepository: mockRepo}

	url := "/devices/switch"
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader("invalid json"))
	ctx := context.WithValue(req.Context(), "email", "	invalid@example.com")
	ctx = context.WithValue(ctx, "role", "admin")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	ctlr.SwitchDeviceMode(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Errorf("expected body to contain 'Invalid request body', got %q", rr.Body.String())
	}
}