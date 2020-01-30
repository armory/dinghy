package server

import "fmt"

type ServerConfig struct {
	Host string `yaml:"host"`
	Port uint32 `yaml:"port"`
	Ssl  Ssl    `yaml:"ssl"`
}

func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type ClientAuthType string

const (
	// No mTLS
	ClientAuthNone ClientAuthType = "none"
	// Client cert verified if provided
	ClientAuthWant ClientAuthType = "want"
	// Client cert required and verified
	ClientAuthNeed ClientAuthType = "need"
	// Any client cert will do
	ClientAuthAny ClientAuthType = "any"
	// Request client cert
	ClientAuthRequest ClientAuthType = "request"
)

type Ssl struct {
	// Enable SSL
	Enabled bool `yaml:"enabled"`
	// Certificate file, can be just a PEM of cert + key or just the cert in which case you'll also need
	// to provide the key file
	CertFile string `yaml:"certFile"`
	// Key file if the cert file doesn't provide it
	KeyFile string `yaml:"keyFile"`
	// Key password if the key is encrypted
	KeyPassword string `yaml:"keyFilePassword"`
	// when using mTLS, CA PEM. If not provided, it will default to the certificate of the server as a CA
	CAcertFile string `yaml:"cacertFile"`
	// Client auth requested (none, want, need, any, request)
	ClientAuth ClientAuthType `yaml:"clientAuth"`
}
