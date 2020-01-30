package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	tls2 "github.com/armory/go-yaml-tools/pkg/tls"
	"io/ioutil"
	"net/http"
)

type Server struct {
	config *ServerConfig
	server *http.Server
}

func NewServer(config *ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

// Start starts the server on the configured port
func (s *Server) Start(router http.Handler) error {
	if !s.config.Ssl.Enabled {
		return s.startHttp(router)
	}
	return s.startTls(router)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) startHttp(router http.Handler) error {
	s.server = &http.Server{
		Addr:    s.config.GetAddr(),
		Handler: router,
	}
	return s.server.ListenAndServe()
}

func (s *Server) startTls(router http.Handler) error {
	tlsConfig, err := s.tlsConfig()
	if err != nil {
		return err
	}

	certMode := s.getClientCertMode()
	if certMode != tls.NoClientCert {
		// With mTLS, we'll parse our PEM to discover CAs with which to validate client certificates
		caFile := s.config.Ssl.CAcertFile
		if caFile == "" {
			// Fall back to cert file - could be a combined PEM (e.g. self signed)
			caFile = s.config.Ssl.CertFile
		} else if err := tls2.CheckFileExists(caFile); err != nil {
			return fmt.Errorf("error with certificate authority file %s: %w", caFile, err)
		}

		// Create a CA certificate pool and add our server certificate
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = certMode
	}

	// Discover the server name based on given certificates
	tlsConfig.BuildNameToCertificate()

	// Create a Server instance to listen on port 8443 with the TLS config
	s.server = &http.Server{
		Addr:      s.config.GetAddr(),
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	// Listen to HTTPS connections with the server certificate and wait
	return s.server.ListenAndServeTLS("", "")
}

func (s *Server) getClientCertMode() tls.ClientAuthType {
	switch s.config.Ssl.ClientAuth {
	case ClientAuthWant:
		return tls.VerifyClientCertIfGiven
	case ClientAuthNeed:
		return tls.RequireAndVerifyClientCert
	case ClientAuthAny:
		return tls.RequireAnyClientCert
	case ClientAuthRequest:
		return tls.RequestClientCert
	default:
		return tls.NoClientCert
	}
}

// tlsConfig prepares the TLS config of the server
// certFile must contain the certificate of the server. It can also contain the private key (optionally encrypted)
// keyFile is needed if the certFile doesn't contain the private key. It can also be encrypted.
func (s *Server) tlsConfig() (*tls.Config, error) {
	c, err := tls2.GetX509KeyPair(s.config.Ssl.CertFile, s.config.Ssl.KeyFile, s.config.Ssl.KeyPassword)
	if err != nil {
		return nil, fmt.Errorf("error with certificate file %s: %w", s.config.Ssl.CertFile, err)
	}
	return &tls.Config{
		Certificates:             []tls.Certificate{c},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
	}, nil
}
