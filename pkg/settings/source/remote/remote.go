package remote

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/dinghy/pkg/util"
	logr "github.com/sirupsen/logrus"
	"net/http"
)

const (
	//RemoteConfigSource is a variable of type string
	RemoteConfigSource = "RemoteSource"
)

//RemoteSource is file source
type RemoteSource struct {
}

//NewRemoteSource creates a source which can handler remote settings
func NewRemoteSource() *RemoteSource {
	remoteConfigSource := new(RemoteSource)
	return remoteConfigSource
}

//LoadSetupSettings load setttings for dinghy start-up
func (lSource *RemoteSource) LoadSetupSettings(*logr.Logger) (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	config.SQL.Enabled = true
	config.SQL.EventLogsOnly = false

	return config, nil
}

//GetSourceName get name of source
func (*RemoteSource) GetSourceName() string {
	return RemoteConfigSource
}

//GetSettings get settings given the key
func (lSource *RemoteSource) GetSettings(r *http.Request, logger *logr.Logger) (*global.Settings, util.PlankClient, error) {
	return nil,nil, nil
}

func (*RemoteSource) BustCacheHandler(w http.ResponseWriter, r *http.Request) {}

func (*RemoteSource) IsMultiTenant() bool {
	return true
}
