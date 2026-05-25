package routes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"mqtt-streaming-server/utils"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	keyfunc "github.com/MicahParks/keyfunc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(db *mongo.Database, mqttClient mqtt.Client) http.Handler {
	mux := http.NewServeMux()
	InitUserRoutes(db, mux)
	InitPhotoRoutes(db, mux)
	InitDeviceRoutes(db, mqttClient, mux)

	// Broker info endpoint - returns the MQTT broker connection info
	mux.HandleFunc("/broker-info", handleBrokerInfo)

	corsHandler := withCORS(mux)

	// Add other middleware here if needed
	return corsHandler
}

// handleBrokerInfo returns the MQTT broker IP and port for client connections
func handleBrokerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the server's local IP address
	ip := getOutboundIP()
	port := "1883" // Default MQTT port

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ip":   ip,
		"port": port,
	})
}

// getOutboundIP gets the preferred outbound IP of this machine
// In Docker, we need to use the host's external IP, not the container IP
func getOutboundIP() string {
	// First, check if MQTT_HOST_IP is set explicitly
	if hostIP := os.Getenv("MQTT_HOST_IP"); hostIP != "" {
		// If it's a hostname (like host.docker.internal), resolve it
		if addrs, err := net.LookupHost(hostIP); err == nil && len(addrs) > 0 {
			return addrs[0]
		}
		// If it's already an IP, return as-is
		if net.ParseIP(hostIP) != nil {
			return hostIP
		}
	}

	// Try to resolve host.docker.internal (works in Docker Desktop)
	addrs, err := net.LookupHost("host.docker.internal")
	if err == nil && len(addrs) > 0 {
		return addrs[0]
	}

	// Fallback: detect outbound IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace * with your domain in production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func noAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No authentication - pass through with placeholder context values
		ctx := context.WithValue(r.Context(), "email", "guest@example.com")
		ctx = context.WithValue(ctx, "role", "user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var (
	jwksOnce   sync.Once
	globalJWKS *keyfunc.JWKS
)

// authMiddleware builds an auth handler from any jwt.Keyfunc, allowing tests to inject a mock.
func authMiddleware(keyfuncFn jwt.Keyfunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.Parse(tokenString, keyfuncFn)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		email, _ := claims["email"].(string)
		if email == "" {
			email, _ = claims["preferred_username"].(string)
		}

		role := "user"
		if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
			if roles, ok := realmAccess["roles"].([]interface{}); ok {
				for _, r := range roles {
					if r == "admin" {
						role = "admin"
						break
					}
				}
			}
		}

		ctx := context.WithValue(r.Context(), "email", email)
		ctx = context.WithValue(ctx, "role", role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withAuth(next http.Handler) http.Handler {
	jwksOnce.Do(func() {
		keycloakURL := os.Getenv("KEYCLOAK_URL")
		if keycloakURL == "" {
			keycloakURL = "http://keycloak:8080"
		}
		realm := os.Getenv("KEYCLOAK_REALM")
		if realm == "" {
			realm = "ss-project"
		}

		jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, realm)
		var client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: utils.GetCACertPool(),
				},
			},
		}
		var err error
		globalJWKS, err = keyfunc.Get(jwksURL, keyfunc.Options{
			Client:          client,
			RefreshInterval: time.Hour,
			RefreshTimeout:  5 * time.Minute,
		})
		if err != nil {
			panic(fmt.Sprintf("failed to initialize JWKS from %s: %v", jwksURL, err))
		}
	})

	return authMiddleware(globalJWKS.Keyfunc, next)
}
