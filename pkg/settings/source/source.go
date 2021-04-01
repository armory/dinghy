package source

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"net/http"
)

//go:generate mockgen -destination=source_configuration_mock.go -package source -source source.go

type SourceConfiguration interface {
	GetSourceName() string
	LoadSetupSettings() (*global.Settings, error)
	GetSettings(r *http.Request) (*global.Settings, error)
}
