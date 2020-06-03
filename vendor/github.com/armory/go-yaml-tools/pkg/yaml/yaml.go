package yaml

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/armory/go-yaml-tools/pkg/secrets"

	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
)

//Resolve takes an array of yaml maps and returns a single map of a merged
//properties.  The order of `ymlTemplates` matters, it should go from lowest
//to highest precendence.
func Resolve(ymlTemplates []map[interface{}]interface{}, envKeyPairs map[string]string) (map[string]interface{}, error) {
	log.Debugf("Using environ %+v\n", envKeyPairs)

	mergedMap := map[interface{}]interface{}{}
	for _, yml := range ymlTemplates {
		if err := mergo.Merge(&mergedMap, yml, mergo.WithOverride); err != nil {
			log.Error(err)
		}
	}

	// unlike other secret engines, the vault config needs to be registered before it can decrypt anything
	vaultCfg := extractVaultConfig(mergedMap)
	if vaultCfg != nil && (secrets.VaultConfig{}) != *vaultCfg {
		if err := secrets.RegisterVaultConfig(*vaultCfg); err != nil {
			log.Errorf("Error registering vault config: %v", err)
		}
	}

	stringMap := convertToStringMap(mergedMap)

	err := subValues(stringMap, stringMap, envKeyPairs)

	if err != nil {
		return nil, err
	}

	return stringMap, nil
}

func extractVaultConfig(m map[interface{}]interface{})  *secrets.VaultConfig {
	if secretsMap, ok := m["secrets"].(map[interface{}]interface{}); ok {
		if vaultmap, ok := secretsMap["vault"].(map[interface{}]interface{}); ok {
			cfg, err := secrets.DecodeVaultConfig(vaultmap)
			if err != nil {
				log.Errorf("Error decoding vault config: %v", err)
				return nil
			}
			return cfg
		}
	}
	return nil
}

func convertToStringMap(m map[interface{}]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	for k, v := range m {
		switch v.(type) {
		case map[interface{}]interface{}:
			stringMap := convertToStringMap(v.(map[interface{}]interface{}))
			newMap[k.(string)] = stringMap
		case []interface{}:
			var collection []interface{}
			for _, vv := range v.([]interface{}) {
				switch vv.(type) {
				case map[interface{}]interface{}:
					collection = append(collection, convertToStringMap(vv.(map[interface{}]interface{})))
				case string:
					collection = append(collection, fmt.Sprintf("%v", vv))
				}
			}
			newMap[k.(string)] = collection

		default:
			newMap[k.(string)] = fmt.Sprintf("%v", v)
		}
	}
	return newMap
}

func subValues(fullMap map[string]interface{}, subMap map[string]interface{}, env map[string]string) error {
	//responsible for finding all variables that need to be substituted
	keepResolving := true
	loops := 0
	re := regexp.MustCompile("\\$\\{(.*?)\\}")
	for keepResolving && loops < len(subMap) {
		loops++
		for k, value := range subMap {
			switch value.(type) {
			case map[string]interface{}:
				err := subValues(fullMap, value.(map[string]interface{}), env)
				if err != nil {
					return err
				}
			case []interface{}:
				sliceMap := make(map[string]interface{})
				for i := 0; i < len(value.([]interface{})); i++ {
					sliceMap[string(i)] = value.([]interface{})[i]
				}
				err := subValues(fullMap, sliceMap, env)
				if err != nil {
					return err
				}
			case string:
				valStr := value.(string)

				secret, wasSecret, err := resolveSecret(valStr)
				if err != nil {
					return err
				}

				if wasSecret {
					subMap[k] = secret
					continue
				}

				keys := re.FindAllStringSubmatch(valStr, -1)
				for _, keyToSub := range keys {
					resolvedValue := resolveSubs(fullMap, keyToSub[1], env)
					subMap[k] = strings.Replace(valStr, "${"+keyToSub[1]+"}", resolvedValue, -1)
				}
			}
		}
	}
	return nil
}

func resolveSecret(valStr string) (string, bool, error) {
	// if the value is a secret resolve it
	if secrets.IsEncryptedSecret(valStr) {
		d, err := secrets.NewDecrypter(context.TODO(), valStr)
		if err != nil {
			return "", true, err
		}

		secret, err := d.Decrypt()
		if err != nil {
			return "", true, err
		}

		return secret, true, nil
	}
	return valStr, false, nil
}

func resolveSubs(m map[string]interface{}, keyToSub string, env map[string]string) string {
	//this function returns array of tuples with their substituted values
	//this handles the case of multiple substitutions in a value
	//baseUrl: ${services.default.protocol}://${services.echo.host}:${services.echo.port}
	keyDefaultSplit := strings.Split(keyToSub, ":")
	subKey, defaultKey := keyDefaultSplit[0], keyDefaultSplit[0]
	if len(keyDefaultSplit) == 2 {
		defaultKey = keyDefaultSplit[1]
	}

	if v := valueFromFlatKey(subKey, m); v != "" {
		defaultKey = v
	} else if v, ok := env[subKey]; ok {
		defaultKey = v
	}

	return defaultKey
}

func valueFromFlatKey(flatKey string, m map[string]interface{}) string {
	keys := strings.Split(flatKey, ".")
	for _, key := range keys {
		switch m[key].(type) {
		case map[string]interface{}:
			m = m[key].(map[string]interface{})
		case string:
			return m[key].(string)
		default:
			continue
		}
	}
	return ""
}
