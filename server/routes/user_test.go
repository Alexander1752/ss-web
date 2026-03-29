package routes_test

import (
	"context"
	"encoding/json"
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

func TestUserController_Register(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      string
		mockSaveReturn error
		expectedStatus int
		expectedUser   *domain.User
	}{
		{
			name:           "successful registration",
			inputBody:      `{"email": "test@example.com", "password": "securepass"}`,
			mockSaveReturn: nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "user already exists",
			inputBody:      `{"email": "test@example.com", "password": "securepass"}`,
			mockSaveReturn: errors.New("user already exists"),
			expectedStatus: http.StatusConflict,
			expectedUser: &domain.User{
				Email: "test@example.com",
			},
		},
		{
			name:           "invalid JSON",
			inputBody:      `invalid-json`,
			mockSaveReturn: nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repository error",
			inputBody:      `{"email": "test@example.com", "password": "securepass"}`,
			mockSaveReturn: errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockUserRepository(ctrl)
			ctlr := routes.UserController{UserRepository: mockRepo}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			if tt.expectedStatus != http.StatusBadRequest {
				mockRepo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(tt.expectedUser, nil)
			}
			if tt.expectedStatus != http.StatusBadRequest && tt.expectedStatus != http.StatusConflict {
				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.mockSaveReturn)
			}

			ctlr.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestUserController_Login(t *testing.T) {
	tests := []struct {
		name             string
		inputBody        string
		mockUser         *domain.User
		mockError        error
		expectedStatus   int
		expectedContains string
	}{
		{
			name:      "successful login",
			inputBody: `{"email": "test@example.com", "password": "password123"}`,
			mockUser: &domain.User{
				Email:    "test@example.com",
				Password: "$2a$12$.OZ5oYXEsFvcaaVh/nmgt.cknGSFzKVlr.wkrzyCl5rgHuAGGkhiS",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:             "invalid JSON",
			inputBody:        `not-a-json`,
			expectedStatus:   http.StatusBadRequest,
			expectedContains: "Invalid request body",
		},
		{
			name:             "user not found",
			inputBody:        `{"email": "missing@example.com", "password": "password123"}`,
			mockUser:         nil,
			mockError:        errors.New("user not found"),
			expectedStatus:   http.StatusUnauthorized,
			expectedContains: "Invalid email or password",
		},
		{
			name:           "repository error",
			inputBody:      `{"email": "test@example.com", "password": "password123"}`,
			mockUser:       nil,
			mockError:      errors.New("db error"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockUserRepository(ctrl)
			ctlr := routes.UserController{UserRepository: mockRepo}

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			if tt.mockUser != nil || tt.mockError != nil {
				mockRepo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(tt.mockUser, tt.mockError)
			}

			ctlr.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedContains, rr.Body.String())
			}
		})
	}
}

func TestUserController_GetProfile(t *testing.T) {
	tests := []struct {
		name             string
		userEmail        string
		mockUser         *domain.User
		mockError        error
		expectedStatus   int
		expectedContains string
	}{
		{
			name:      "successful profile fetch",
			userEmail: "test@example.com",
			mockUser: &domain.User{
				Email: "test@example.com",
			},
			expectedStatus:   http.StatusOK,
			expectedContains: "test@example.com",
		},
		{
			name:             "user not found",
			userEmail:        "missing@example.com",
			mockUser:         nil,
			mockError:        errors.New("user not found"),
			expectedStatus:   http.StatusNotFound,
			expectedContains: "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockUserRepository(ctrl)
			ctlr := routes.UserController{UserRepository: mockRepo}

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			ctx := context.WithValue(req.Context(), "email", tt.userEmail)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			if tt.userEmail != "" {
				mockRepo.EXPECT().FindByEmail(gomock.Any(), tt.userEmail).Return(tt.mockUser, tt.mockError)
			}

			ctlr.GetProfile(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedContains, rr.Body.String())
			}
		})
	}
}

func TestUserController_GetProfile_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockUserRepository(ctrl)
	ctlr := routes.UserController{UserRepository: mockRepo}

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rr := httptest.NewRecorder()

	ctlr.GetProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Email not found in context") {
		t.Errorf("expected body to contain 'Email not found in context', got %q", rr.Body.String())
	}
}

func TestUserController_GetProfile_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockUserRepository(ctrl)
	ctlr := routes.UserController{UserRepository: mockRepo}

	req := httptest.NewRequest(http.MethodPost, "/profile", nil)
	rr := httptest.NewRecorder()

	ctlr.GetProfile(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestUserController_Register_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockUserRepository(ctrl)
	ctlr := routes.UserController{UserRepository: mockRepo}

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rr := httptest.NewRecorder()

	ctlr.Register(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestUserController_Login_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockUserRepository(ctrl)
	ctlr := routes.UserController{UserRepository: mockRepo}

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()

	ctlr.Login(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestUserController_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockUserRepository(ctrl)
	ctlr := routes.UserController{UserRepository: mockRepo}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@example.com", "password": "wrongpassword"}`))
	ctx := context.WithValue(req.Context(), "email", "example@example.com")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mockRepo.EXPECT().
		FindByEmail(gomock.Any(), "example@example.com").
		Return(&domain.User{
			Email:    "example@example.com",
			Password: "$2a$12$OZ5oYXEsFvcaaVh/nmgt.cknGSFzKVlr.wkrzyCl5rgHuAGGkhiS",
		}, nil)
	ctlr.Login(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Errorf("expected body to contain 'Invalid email or password', got %q", rr.Body.String())
	}
}

func FuzzUserController_Register(f *testing.F) {
	seedInputs := []string{
		`{"email": "user@example.com", "password": "pass1234"}`,
		`{"email": "", "password": ""}`,
		`{"email": "a@b.c", "password": "short"}`,
		`not-json`,
		`{"email": "incomplete`,
	}

	for _, input := range seedInputs {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockUserRepository(ctrl)
		ctlr := routes.UserController{UserRepository: mockRepo}

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(input))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		var parsed struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.Unmarshal([]byte(input), &parsed); err == nil {
			mockRepo.EXPECT().
				FindByEmail(gomock.Any(), parsed.Email).
				Return(nil, nil).
				AnyTimes()

			mockRepo.EXPECT().
				Save(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
		} else {
			mockRepo.EXPECT().
				FindByEmail(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()
			mockRepo.EXPECT().
				Save(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
		}

		ctlr.Register(rr, req)

		if rr.Code < 100 || rr.Code > 599 {
			t.Errorf("unexpected status code: %d for input: %q", rr.Code, input)
		}
	})
}
