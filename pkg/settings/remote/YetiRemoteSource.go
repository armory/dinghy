package remote

import (
	"github.com/armory-io/dinghy/pkg/settings/remote/yeti"
	"github.com/armory/dinghy/pkg/debug"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/plank/v4"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	//RemoteConfigSource is a variable of type string
	RemoteConfigSource = "YetiRemoteSource"
)

//RemoteSource is file source
type YetiRemoteSource struct {
	cache      *cache.Cache
	yetiClient yeti.Client
}

//NewRemoteSource creates a source which can handler remote settings
func NewRemoteSource() *YetiRemoteSource {
	remoteConfigSource := new(YetiRemoteSource)
	remoteConfigSource.cache = cache.New(1*time.Hour, 10*time.Minute)
	return remoteConfigSource
}

func (rSource *YetiRemoteSource) SetYetiEndpoint(endpoint string) {
	rSource.yetiClient = yeti.NewYetiClient(endpoint)
}

//LoadSetupSettings load setttings for dinghy start-up
func (rSource *YetiRemoteSource) LoadSetupSettings(logr *log.Logger) (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	//config.SQL.Enabled = true
	config.SQL.EventLogsOnly = false

	return config, nil
}

//GetSourceName get name of source
func (*YetiRemoteSource) GetSourceName() string {
	return RemoteConfigSource
}

//GetSettings get settings given the key
//We tied this method to the request object because we couldn't prioritize figuring out the best way to obfuscate the use
//of evn/org ID in the OSS interface. It's not ideal but it works, we should just note we may need to refactor later if needed.

func (rSource *YetiRemoteSource) GetSettings(r *http.Request, logr *log.Logger) (*global.Settings, util.PlankClient, error) {
	envId := r.Header.Get("X-Environment-ID")
	orgId := r.Header.Get("X-Organization-ID")
	settings, found := rSource.cache.Get(envId)
	if found {
		cachedSettings := settings.(*global.Settings)
		plankClinet := setupPlankClient(cachedSettings, logr)
		return cachedSettings, plankClinet, nil
	} else {
		//get from remote source
		var err error
		settings, err = rSource.yetiClient.GetSettings(envId, orgId)
		if err != nil {
			log.Errorf("Unable to retrieve config from Yeti")
			return nil, nil, err
		}
		rSource.cache.Set(envId, settings, cache.DefaultExpiration)
		plankClinet := setupPlankClient(settings.(*global.Settings), logr)
		return settings.(*global.Settings), plankClinet, nil
	}
}

//Delete settings for a given the key
func (rSource *YetiRemoteSource) BustCacheHandler(w http.ResponseWriter, r *http.Request) {
	envId := r.Header.Get("X-Environment-ID")
	if envId == "" {
		log.Errorf("Must provide X-Environment-Id header")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		rSource.cache.Delete(envId)
		w.WriteHeader(http.StatusOK)
	}
}

func (rSource *YetiRemoteSource) IsMultiTenant() bool{
	return true
}

func setupPlankClient(settings *global.Settings, logr *log.Logger) *plank.Client {
	var httpClient *http.Client
	if logr.Level == log.DebugLevel {
		httpClient = debug.NewInterceptorHttpClient(logr, &settings.Http, true)
	} else {
		httpClient = settings.Http.NewClient()
	}
	client := plank.New(plank.WithClient(httpClient),
		plank.WithFiatUser(settings.SpinnakerSupplied.Fiat.AuthUser))

	// Update the base URLs based on config
	client.URLs["gate"] = settings.SpinnakerSupplied.Gate.BaseURL
	client.UseGateEndpoints()
	client.EnableArmoryEndpoints()
	return client
}

