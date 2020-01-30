package config

// Settings mirrors Spinnaker's yaml files.
type Settings struct {
	Services Services `json:"services,omitempty" mapstructure:"services"`
}

// Services within Spinnaker.
type Services struct {
	Fiat    Service `json:"fiat,omitempty" mapstructure:"fiat"`
	Front50 Front50 `json:"front50,omitempty" mapstructure:"front50"`
}

// Front50 service settings.
type Front50 struct {
	Service
	StorageBucket string `json:"storage_bucket,omitempty" mapstructure:"storage_bucket"`
	StoragePrefix string `json:"rootFolder,omitempty" mapstructure:"rootFolder"`
	S3            struct {
		Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
	} `json:"s3,omitempty" mapstructure:"s3"`
	GCS struct {
		Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
	} `json:"gcs,omitempty" mapstructure:"gcs"`
}

// Service such as Fiat, Orca, Deck, Gate, etc.
type Service struct {
	Enabled bool   `json:"enabled,omitempty" mapstructure:"enabled"`
	BaseURL string `json:"baseUrl,omitempty" mapstructure:"baseUrl"`
}
