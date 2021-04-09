package yeti

import (
	"encoding/json"
	"fmt"
	ossSettings "github.com/armory/dinghy/pkg/settings/global"
	"io/ioutil"
	"net/http"
)

type Client struct {
	YetiUrl     string
}

type SettingsResp struct {
	Id           string                 `json:"id,omitempty" yaml:"id"`
	OrgId        string                 `json:"orgId,omitempty" yaml:"orgId"`
	EnvId        string                 `json:"envId,omitempty" yaml:"envId"`
	ResourceType string                 `json:"resourceType,omitempty" yaml:"resourceType"`
	ResourceKind string                 `json:"resourceKind,omitempty" yaml:"resourceKind"`
	CreatedTs    string                 `json:"createdTs,omitempty" yaml:"createdTs"`
	UpdatedTS    string                 `json:"updatedTs,omitempty" yaml:"updatedTs"`
	Dinghy       map[string]interface{} `json:"dinghy,omitempty" yaml:"dinghy"`
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	HttpClient HTTPClient
)

func init() {
	HttpClient = &http.Client{}
}

func NewYetiClient(endpoint string) Client {
	client := Client{YetiUrl: endpoint}
	return client
}

func (y *Client) GetSettings(environmentId, organizationId string) (*ossSettings.Settings, error) {
	//Make the call to Yeti
	endpoint := fmt.Sprintf("%s/api/v1/environments/%s/armory/dinghy", y.YetiUrl, environmentId)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return &ossSettings.Settings{}, fmt.Errorf("unable to create new HTTP Request for getting remote Dinghy settings from Yeti. Error: %s", err)
	}
	req.Header.Add("X-Armory-Organization-ID", organizationId)
	resp, err := HttpClient.Do(req)
	if err != nil {
		return &ossSettings.Settings{}, fmt.Errorf("unable to successfully complete request to yeti to retrieve remote Dinghy settings from Yeti. Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ossSettings.Settings{}, fmt.Errorf("unable to read HTTP response from Yeti when getting remote Dinghy settings. Error: %s", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Unmarshal Yeti response into the SettingsResp struct
		obj := SettingsResp{}
		err = json.Unmarshal(body, &obj)
		if err != nil {
			return &ossSettings.Settings{}, fmt.Errorf("unable to unmarshal Yeti response to JSON when getting remote Dinghy settins. Error: %s", err)
		}

		//In order to isolate the Dinghy config we need to marshal the config back into JSON and then unmarshal it into the correct struct types
		jsonString, err := json.Marshal(obj.Dinghy)
		if err != nil {
			return &ossSettings.Settings{}, fmt.Errorf("unable to remarshal Dinghy config to JSON. Error: %s", err)
		}
		ossSettingsMap := &ossSettings.Settings{}
		err = json.Unmarshal(jsonString, &ossSettingsMap)
		if err != nil {
			return &ossSettings.Settings{}, fmt.Errorf("unable to unmarshal OSS Dinghy settings to JSON. Error: %s", err)
		}
		/*extSettingsMap := settings.ExtSettings{}
		err = json.Unmarshal(jsonString, &extSettingsMap)
		if err != nil {
			return ossSettings.Settings{}, fmt.Errorf("unable to unmarshal internal Dinghy settings to JSON. Error: %s", err)
		}
		extSettingsMap.Settings = &ossSettingsMap*/
		return ossSettingsMap, nil
	} else if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		return &ossSettings.Settings{}, fmt.Errorf("non-OK HTTP status when requesting remote Dinghy settings from Yeti. Status Code %d, \n Response Body: %s", resp.StatusCode, string(body))
	} else {
		// If status falls outside the range of 200 - 599 then return an error.
		return &ossSettings.Settings{}, fmt.Errorf("non-OK HTTP status when requesting remote Dinghy settings from Yeti. Status Code %d", resp.StatusCode)
	}

}
