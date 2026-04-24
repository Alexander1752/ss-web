package utils

import (
	"crypto/x509"
	"fmt"
	"os"
)

func GetCACertPool() *x509.CertPool {
	caCert, err := os.ReadFile("/certs/ca.crt")
	if err != nil {
		panic(fmt.Errorf("failed to read CA certificate: %w", err))
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	return certPool
}
