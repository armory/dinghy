package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"io/ioutil"
	"net/http"
	"os"
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
		} else if err := checkFileExists(caFile); err != nil {
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

func checkFileExists(filename string) error {
	if secrets.IsEncryptedSecret(filename) {
		d, err := secrets.NewDecrypter(context.TODO(), filename)
		if err != nil {
			return err
		}
		if !d.IsFile() {
			return errors.New("no file referenced, use encryptedFile")
		}
		filename, err = d.Decrypt()
		if err != nil {
			return err
		}
	}
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return err
	}
	return nil
}

// tlsConfig prepares the TLS config of the server
// certFile must contain the certificate of the server. It can also contain the private key (optionally encrypted)
// keyFile is needed if the certFile doesn't contain the private key. It can also be encrypted.
func (s *Server) tlsConfig() (*tls.Config, error) {
	if err := checkFileExists(s.config.Ssl.CertFile); err != nil {
		return nil, fmt.Errorf("error with certificate file %s: %w", s.config.Ssl.CertFile, err)
	}

	b, err := ioutil.ReadFile(s.config.Ssl.CertFile)
	if err != nil {
		return nil, err
	}

	pemBlocks, pkey, err := s.readAndDecryptPEM(b)

	// If private key not in the cert file, we look for it in the key file
	if pkey == nil {
		pkey, err = s.getPrivateKey()
		if err != nil {
			return nil, err
		}
	}

	c, _ := tls.X509KeyPair(pem.EncodeToMemory(pemBlocks[0]), pkey)

	return &tls.Config{
		Certificates:             []tls.Certificate{c},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
	}, nil
}

// getPrivateKey attempts to load and decrypt the private key if needed
func (s *Server) getPrivateKey() ([]byte, error) {
	if err := checkFileExists(s.config.Ssl.KeyFile); err != nil {
		return nil, fmt.Errorf("error with key file %s: %w", s.config.Ssl.KeyFile, err)
	}
	b, err := ioutil.ReadFile(s.config.Ssl.KeyFile)
	if err != nil {
		return nil, err
	}
	_, pkey, err := s.readAndDecryptPEM(b)
	return pkey, err
}

// readAndDecryptPEM reads PEM data and attempts to decrypt if a private key is found encrypted
// using ssl.keyPassword provided in the config
func (s *Server) readAndDecryptPEM(data []byte) ([]*pem.Block, []byte, error) {
	var pemBlocks []*pem.Block
	var v *pem.Block
	var pkey []byte

	for {
		v, data = pem.Decode(data)
		if v == nil {
			break
		}
		if v.Type == "RSA PRIVATE KEY" {
			if x509.IsEncryptedPEMBlock(v) {
				pass, err := s.getKeyPassword()
				if err != nil {
					return nil, nil, err
				}
				pkey, _ = x509.DecryptPEMBlock(v, []byte(pass))
				pkey = pem.EncodeToMemory(&pem.Block{
					Type:  v.Type,
					Bytes: pkey,
				})
			} else {
				pkey = pem.EncodeToMemory(v)
			}
		} else {
			pemBlocks = append(pemBlocks, v)
		}
	}
	return pemBlocks, pkey, nil
}

func (s *Server) getKeyPassword() (string, error) {
	if secrets.IsEncryptedSecret(s.config.Ssl.KeyPassword) {
		d, err := secrets.NewDecrypter(context.TODO(), s.config.Ssl.KeyPassword)
		if err != nil {
			return "", err
		}
		return d.Decrypt()
	}
	if s.config.Ssl.KeyPassword == "" {
		return "", fmt.Errorf("encrypted pem found but no password provided")
	}
	return s.config.Ssl.KeyPassword, nil
}
