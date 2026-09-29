package platform

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"go.temporal.io/sdk/client"
)

func DialTemporal() (client.Client, error) {
	options := client.Options{HostPort: Env("TEMPORAL_ADDRESS", "localhost:7233"), Namespace: Env("TEMPORAL_NAMESPACE", "default")}
	certPath, keyPath := os.Getenv("TEMPORAL_TLS_CERT"), os.Getenv("TEMPORAL_TLS_KEY")
	caPath := os.Getenv("TEMPORAL_TLS_CA")
	if (certPath == "") != (keyPath == "") {
		return nil, errors.New("TEMPORAL_TLS_CERT and TEMPORAL_TLS_KEY must be set together")
	}
	if certPath != "" || caPath != "" || os.Getenv("TEMPORAL_TLS_ENABLED") == "true" {
		config := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: os.Getenv("TEMPORAL_TLS_SERVER_NAME")}
		if certPath != "" {
			certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, err
			}
			config.Certificates = []tls.Certificate{certificate}
		}
		if caPath != "" {
			roots, err := x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
			pem, err := os.ReadFile(caPath)
			if err != nil {
				return nil, err
			}
			if !roots.AppendCertsFromPEM(pem) {
				return nil, errors.New("invalid Temporal CA certificate")
			}
			config.RootCAs = roots
		}
		options.ConnectionOptions.TLS = config
	}
	return client.Dial(options)
}
