package spinnaker

import (
	"crypto/tls"
	"crypto/x509"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"

	"github.com/armory-io/dinghy/pkg/settings"
)

var defaultClient *http.Client

func init() {
	c, err := newX509Client()
	if err != nil {
		log.Error("Could not create x509 client.", err)
	}
	defaultClient = c
}

func newX509Client() (*http.Client, error) {
	var c http.Client
	log.Info("Configuring TLS Spinnaker Client with certificate")
	cert, err := tls.LoadX509KeyPair(settings.S.CertPath, settings.S.CertPath)
	if err != nil {
		return nil, err
	}
	clientCA, err := ioutil.ReadFile(settings.S.CertPath)
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCA)
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		InsecureSkipVerify:       true,
	}
	c.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return &c, nil
}
