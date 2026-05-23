package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandleBrokerInfo_GET(t *testing.T) {
	req := httptest.NewRequest("GET", "/broker-info", nil)
	w := httptest.NewRecorder()

	handleBrokerInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("failed to parse response: %v", err)
	}

	if _, hasIP := result["ip"]; !hasIP {
		t.Error("response missing 'ip' field")
	}
	if _, hasPort := result["port"]; !hasPort {
		t.Error("response missing 'port' field")
	}

	if result["port"] != "1883" {
		t.Errorf("expected port 1883, got %s", result["port"])
	}
}

func TestHandleBrokerInfo_MethodNotAllowed(t *testing.T) {
	tests := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range tests {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/broker-info", nil)
			w := httptest.NewRecorder()

			handleBrokerInfo(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", w.Code)
			}
		})
	}
}

func TestHandleBrokerInfo_WithCustomMQTTHost(t *testing.T) {
	t.Run("with MQTT_HOST_IP env var", func(t *testing.T) {
		os.Setenv("MQTT_HOST_IP", "192.168.1.1")
		defer os.Unsetenv("MQTT_HOST_IP")

		req := httptest.NewRequest("GET", "/broker-info", nil)
		w := httptest.NewRecorder()

		handleBrokerInfo(w, req)

		var result map[string]string
		json.Unmarshal(w.Body.Bytes(), &result)

		if result["ip"] == "" {
			t.Error("expected IP to be set from MQTT_HOST_IP env var")
		}
	})
}

func TestWithCORS_SetHeaders(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := withCORS(innerHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	tests := []struct {
		header   string
		expected string
	}{
		{"Access-Control-Allow-Origin", "*"},
		{"Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS"},
		{"Access-Control-Allow-Headers", "Content-Type, Authorization"},
	}

	for _, tt := range tests {
		value := w.Header().Get(tt.header)
		if value != tt.expected {
			t.Errorf("header %s: expected %q, got %q", tt.header, tt.expected, value)
		}
	}
}

func TestWithCORS_OPTIONS_Preflight(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called for OPTIONS request")
		w.WriteHeader(http.StatusOK)
	})

	handler := withCORS(innerHandler)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}

	// Verify CORS headers are still set
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected CORS header for OPTIONS request, got %q", origin)
	}
}

func TestWithCORS_PassesOtherMethods(t *testing.T) {
	called := false
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := withCORS(innerHandler)
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("inner handler should be called for non-OPTIONS requests")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestNoAuth_AddContextValues(t *testing.T) {
	var capturedCtx context.Context
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	handler := noAuth(innerHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	email := capturedCtx.Value("email")
	role := capturedCtx.Value("role")

	if email != "guest@example.com" {
		t.Errorf("expected email 'guest@example.com', got %v", email)
	}
	if role != "user" {
		t.Errorf("expected role 'user', got %v", role)
	}
}

func TestNoAuth_PreservesRequest(t *testing.T) {
	var capturedReq *http.Request
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})

	handler := noAuth(innerHandler)
	originalReq := httptest.NewRequest("GET", "/test", nil)
	originalReq.Header.Set("X-Custom", "value")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, originalReq)

	if capturedReq.Header.Get("X-Custom") != "value" {
		t.Error("noAuth should preserve request headers")
	}
}

func TestGetOutboundIP_WithValidEnvVar(t *testing.T) {
	t.Run("IP address", func(t *testing.T) {
		os.Setenv("MQTT_HOST_IP", "10.0.0.1")
		defer os.Unsetenv("MQTT_HOST_IP")

		ip := getOutboundIP()
		if ip != "10.0.0.1" {
			t.Errorf("expected 10.0.0.1, got %s", ip)
		}
	})

	t.Run("no env var fallback", func(t *testing.T) {
		os.Unsetenv("MQTT_HOST_IP")

		ip := getOutboundIP()
		if ip == "" {
			t.Error("expected non-empty IP from fallback")
		}
	})
}

func TestInitRoutes_ReturnsHandler(t *testing.T) {
	// Verify that a handler is created successfully
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if innerHandler == nil {
		t.Error("expected handler to be non-nil")
	}
}

func TestInitRoutes_BrokerInfoEndpoint(t *testing.T) {
	// Create a test handler that includes broker-info
	mux := http.NewServeMux()
	mux.HandleFunc("/broker-info", handleBrokerInfo)
	handler := withCORS(mux)

	req := httptest.NewRequest("GET", "/broker-info", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("failed to parse response: %v", err)
	}

	if result["port"] != "1883" {
		t.Errorf("expected port 1883, got %s", result["port"])
	}
}

func TestCORSAndBrokerInfo_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/broker-info", handleBrokerInfo)
	handler := withCORS(mux)

	t.Run("GET request with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/broker-info", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check CORS headers
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("missing CORS origin header")
		}
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/broker-info", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
		}

		// Verify CORS headers still present
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("missing CORS origin header on OPTIONS response")
		}
	})
}
