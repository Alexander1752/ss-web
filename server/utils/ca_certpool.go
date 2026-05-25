package utils

import (
	"crypto/x509"
	"os"
)

func GetCACertPool() *x509.CertPool {
	caCert, err := os.ReadFile("/certs/ca.crt")
	if err != nil {
		return nil
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	return certPool
}
