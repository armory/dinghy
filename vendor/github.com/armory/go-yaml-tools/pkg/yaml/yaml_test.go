package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func check(t *testing.T, e error) {
	if e != nil {
		t.Errorf("error when testing: %s", e)
	}
}

func TestSubValues(t *testing.T) {
	m := map[string]interface{}{
		"mock": map[string]interface{}{
			"somekey": "${mock.flat.otherkey.value}",
			"flat": map[string]interface{}{
				"otherkey": map[string]interface{}{
					"value": "mockReplaceValue",
				},
			},
		},
	}

	subValues(m, m, nil)
	testValue := m["mock"].(map[string]interface{})["somekey"]
	assert.Equal(t, "mockReplaceValue", testValue)
}

func TestResolveSubs(t *testing.T) {
	m := map[string]interface{}{
		"mock": map[string]interface{}{
			"flat": map[string]interface{}{
				"otherkey": map[string]interface{}{
					"value": "mockValue",
				},
			},
		},
	}
	str := resolveSubs(m, "mock.flat.otherkey.value", nil)
	assert.Equal(t, "mockValue", str)
}

func readTestFixtures(t *testing.T, fileName string) map[interface{}]interface{} {
	wd, _ := os.Getwd()
	spinnakerYml := fmt.Sprintf("%s/../../test/%s", wd, fileName)
	f, err := os.Open(spinnakerYml)
	check(t, err)
	s, err := ioutil.ReadAll(f)
	check(t, err)

	any := map[interface{}]interface{}{}
	err = yaml.Unmarshal(s, &any)
	check(t, err)

	return any
}

func TestResolver(t *testing.T) {

	fileNames := []string{
		"spinnaker.yml",
		"spinnaker-armory.yml",
		"spinnaker-local.yml",
	}

	ymlMaps := []map[interface{}]interface{}{}
	for _, f := range fileNames {
		ymlMaps = append(ymlMaps, readTestFixtures(t, f))
	}
	envKeyPairs := map[string]string{
		"SPINNAKER_AWS_ENABLED": "true",
		"DEFAULT_DNS_NAME":      "mockdns.com",
		"REDIS_HOST":            "redishost.com",
	}

	resolved, err := Resolve(ymlMaps, envKeyPairs)
	if err != nil {
		t.Error(err)
	}
	//simple replace
	host := resolved["services"].(map[string]interface{})["rosco"].(map[string]interface{})["host"]
	assert.Equal(t, "mockdns.com", host)

	providers := resolved["providers"].(map[string]interface{})
	services := resolved["services"].(map[string]interface{})
	google := providers["google"].(map[string]interface{})
	googleEnabled := google["enabled"]
	assert.Equal(t, "false", googleEnabled)

	//default when no ENV var is present
	defaultRegion := providers["aws"].(map[string]interface{})["defaultRegion"]
	assert.Equal(t, "us-east-1", defaultRegion)

	//more complex substitution with urls
	fiatURL := services["fiat"].(map[string]interface{})["baseUrl"]
	assert.Equal(t, "http://mockdns.com:7003", fiatURL)

	//empty url
	project := google["primaryCredentials"].(map[string]interface{})["project"]
	assert.Equal(t, "", project)
}
