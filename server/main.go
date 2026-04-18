package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/otiai10/gosseract/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mqtt-streaming-server/broker"
	"mqtt-streaming-server/routes"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := fmt.Sprintf("mongodb://%s:%s@mongo-db:27017/?authSource=admin", os.Getenv("MONGO_INITDB_ROOT_USERNAME"), os.Getenv("MONGO_INITDB_ROOT_PASSWORD"))
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Failed to connect to MongoDB:", err)
		panic(err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	db := mongoClient.Database("mqtt-streaming-server")

	fmt.Println("Connected to MongoDB!")

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ocrClient := gosseract.NewClient()
	ocrClient.SetLanguage("eng", "ron")
	defer ocrClient.Close()
	brokerHandler := broker.NewBrokerHandler(db, ocrClient)

	caCert, err := os.ReadFile("/certs/ca.crt")
	if err != nil {
		panic(fmt.Errorf("failed to read MQTT CA certificate: %w", err))
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		panic("failed to append MQTT CA certificate")
	}

	clientCert, err := tls.LoadX509KeyPair("/certs/client.crt", "/certs/client.key")
	if err != nil {
		panic(fmt.Errorf("failed to load MQTT client certificate: %w", err))
	}

	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
		ServerName:   os.Getenv("MQTT_HOST"),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker("ssl://" + os.Getenv("MQTT_HOST") + ":8883")
	opts.SetClientID("web")
	opts.SetTLSConfig(tlsConfig)

	// Start the connection
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscribe to images topic
	if token := client.Subscribe("ssproject/images/#", 0, brokerHandler.HandlePhoto); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := client.Subscribe("register/#", 0, brokerHandler.RegisterDevice); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := client.Subscribe("device/id/#", 0, brokerHandler.DisconnectDevice); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// Initialize user routes
	handler := routes.InitRoutes(db, client)

	go func() {
		fmt.Println("Starting HTTPS server on port 5000...")
		if err := http.ListenAndServeTLS(":5000", "/certs/server.crt", "/certs/server.key", handler); err != nil {
			panic(err)
		}
	}()

	<-c
}
