package spinnaker

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
)

// Pipeline is the structure used by spinnaker
type Pipeline map[string]interface{}

// Lock is embetted in the pipeline if it should be disabled in the UI.
type Lock struct {
	UI            bool `json:"ui"`
	AllowUnlockUI bool `json:"allowUnlockUi"`
}

// Lock disables the pipeline from being edited from the Spinnaker UI.
func (p Pipeline) Lock() {
	p["locked"] = Lock{UI: true, AllowUnlockUI: true}
}

// Name returns the name of the pipeline.
func (p Pipeline) Name() string {
	val, exists := p["name"]
	if exists {
		name := val.(string)
		return name
	}
	return ""
}

// Application returns the app name of the pipeline.
func (p Pipeline) Application() string {
	val, exists := p["application"]
	if exists {
		application := val.(string)
		return application
	}
	return ""
}

// UpdatePipelines posts pipelines to Spinnaker.
func UpdatePipelines(p []Pipeline) (err error) {
	if len(p) == 0 {
		return
	}
	app := p[0].Application()
	ids, _ := pipelineIDs(app)
	log.Info("Found pipelines for ", app, ": ", ids)
	for _, pipeline := range p {
		if id, exists := ids[pipeline.Name()]; exists {
			pipeline["id"] = id
		}
		log.Info("Updating pipeline: " + pipeline.Name())
		if settings.AutoLockPipelines {
			log.Info("Locking pipeline")
			pipeline.Lock()
		}
		err := updatePipeline(pipeline)
		if err != nil {
			log.Error("Could not post pipeline to Spinnaker ", err)
		}
	}
	return
}

func updatePipeline(p Pipeline) error {
	b, err := json.Marshal(p)
	if err != nil {
		log.Error("Could not marshal pipeline ", err)
		return err
	}
	log.Info("Posing pipeline to Spinnaker: ", string(b))
	c, err := newX509Client()
	if err != nil {
		log.Error("Could not make x509 http client ", err)
		return err
	}
	url := fmt.Sprintf(`%s/pipelines`, settings.SpinnakerAPIURL)
	resp, err := c.Post(url, "application/json", strings.NewReader(string(b)))
	for retry := 0; retry < 10 && resp.StatusCode > 399 && err != nil; retry++ {
		log.Error("Couldn't post to Spinnaker: ", resp.StatusCode, " - ", err)
		time.Sleep(time.Duration(retry*200) * time.Millisecond)
		log.Info("Trying again: Posting pipeline to Spinnaker...")
		resp, err = c.Post(url, "application/json", strings.NewReader(string(b)))
	}
	if resp.StatusCode > 399 {
		log.Error("Could not contact Spinnaker to update pipeline.")
		return fmt.Errorf(`spinnaker returned %d`, resp.StatusCode)
	}
	log.Info("Successfully posted pipeline to Spinnaker.")
	return nil
}

// PipelineIDs returns the pipeline IDs keyed by name for an application.
func pipelineIDs(app string) (map[string]string, error) {
	ids := map[string]string{}
	c, err := newX509Client()
	if err != nil {
		log.Error("Could not make x509 http client ", err)
		return ids, err
	}
	url := fmt.Sprintf("/applications/%s/pipelineConfigs", app)
	resp, err := c.Get(url)
	for retry := 0; retry < 10 && resp.StatusCode > 399 && err != nil; retry++ {
		log.Error("Couldn't post to Spinnaker: ", resp.StatusCode, " - ", err)
		time.Sleep(time.Duration(retry*200) * time.Millisecond)
		log.Info("Trying again: Posting pipeline to Spinnaker...")
		resp, err = c.Get(url)
	}
	if err != nil {
		return ids, err
	}
	if resp.StatusCode != http.StatusOK {
		return ids, errors.New("Unable to fetch pipeline list: " + strconv.Itoa(resp.StatusCode))
	}
	type pipeline struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var pipelines []pipeline
	util.ReadJSON(resp.Body, &pipelines)
	for _, p := range pipelines {
		ids[p.Name] = p.ID
	}
	return ids, nil
}

func newX509Client() (*http.Client, error) {
	var c http.Client
	log.Debug("Configuring TLS with certificate")
	cert, err := tls.LoadX509KeyPair(settings.CertPath, settings.CertPath)
	if err != nil {
		return nil, err
	}
	clientCA, err := ioutil.ReadFile(settings.CertPath)
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
