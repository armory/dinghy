package local

import (
	"github.com/armory/dinghy/pkg/debug"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/plank/v4"
	logr "github.com/sirupsen/logrus"
	"net/http"
)

const (
	//LocalConfigSource is a variable of type string
	LocalConfigSource = "LocalSource"
)

//LocalSource is file source
type LocalSource struct {
	Configs *global.Settings
	Client util.PlankClient
}

//NewLocalSource creates a source which can handler local settings
func NewLocalSource() *LocalSource {
	localConfigSource := new(LocalSource)
	return localConfigSource
}

//LoadSetupSettings load setttings for dinghy start-up
func (lSource *LocalSource) LoadSetupSettings(logr *logr.Logger) (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	lSource.Configs = config
	lSource.Client = setupPlankClient(config, logr)

	return config, nil
}

func (*LocalSource) GetSourceName() string {
	return LocalConfigSource
}

//GetSettings get settings given the key
func (lSource *LocalSource) GetSettings(r *http.Request, logr *logr.Logger) (*global.Settings, util.PlankClient, error) {
	return lSource.Configs, lSource.Client, nil
}

func (*LocalSource) BustCacheHandler(w http.ResponseWriter, r *http.Request) {}

func setupPlankClient(settings *global.Settings, log *logr.Logger) *plank.Client {
	var httpClient *http.Client
	if log.Level == logr.DebugLevel {
		httpClient = debug.NewInterceptorHttpClient(log, &settings.Http, true)
	} else {
		httpClient = settings.Http.NewClient()
	}
	client := plank.New(plank.WithClient(httpClient),
		plank.WithFiatUser(settings.SpinnakerSupplied.Fiat.AuthUser))

	// Update the base URLs based on config
	client.URLs["orca"] = settings.SpinnakerSupplied.Orca.BaseURL
	client.URLs["front50"] = settings.SpinnakerSupplied.Front50.BaseURL
	client.URLs["gate"] = settings.SpinnakerSupplied.Gate.BaseURL
	return client
}

func (*LocalSource) IsMultiTenant() bool {
	return false
}
