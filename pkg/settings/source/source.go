package source

import "github.com/armory/dinghy/pkg/settings/global"

//go:generate mockgen -destination=source_configuration_mock.go -package source -source source.go

type SourceConfiguration interface {
	GetSourceName() string
	LoadSetupSettings() (*global.Settings, error)
	GetSettings(key string) (*global.Settings, error)
}
