package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

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

func withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the JWT token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):] // Remove "Bearer " prefix
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract email from token claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		email, ok := claims["email"].(string)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		role, ok := claims["role"].(string)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		// Store the email in the request context
		ctx := context.WithValue(r.Context(), "email", email)
		ctx = context.WithValue(ctx, "role", role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
