package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	tls2 "github.com/armory/go-yaml-tools/pkg/tls"
	"io/ioutil"
)

type Config struct {
	CacertFile        string `yaml:"cacertFile"`
	ClientCertFile    string `yaml:"clientCertFile"`
	ClientKeyFile     string `yaml:"clientKeyFile"`
	ClientKeyPassword string `yaml:"clientKeyPassword"`

	tlsConf *tls.Config
}

func (c *Config) Init() error {
	var caCertPool *x509.CertPool
	var clientCert tls.Certificate
	var err error
	tlsConf := tls.Config{}
	bTls := false

	if c.CacertFile != "" {
		if err := tls2.CheckFileExists(c.CacertFile); err != nil {
			return fmt.Errorf("unable to find CA certificate: %w", err)
		}
		cd, err := ioutil.ReadFile(c.CacertFile)
		if err != nil {
			return fmt.Errorf("unable to load CA certificate: %w", err)
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(cd)
		tlsConf.RootCAs = caCertPool
		bTls = true
	}

	if c.ClientCertFile != "" {
		clientCert, err = tls2.GetX509KeyPair(c.ClientCertFile, c.ClientKeyFile, c.ClientKeyPassword)
		if err != nil {
			return fmt.Errorf("unable to load client certificate: %w", err)
		}
		tlsConf.Certificates = []tls.Certificate{clientCert}
		bTls = true
	}
	if bTls {
		c.tlsConf = &tlsConf
	}
	return nil
}
