package yaml

import (
	"fmt"
	"regexp"
	"strings"

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
	stringMap := convertToStringMap(mergedMap)

	subValues(stringMap, stringMap, envKeyPairs)
	return stringMap, nil
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

func subValues(fullMap map[string]interface{}, subMap map[string]interface{}, env map[string]string) {
	//responsible for finding all variables that need to be substituted
	keepResolving := true
	loops := 0
	re := regexp.MustCompile("\\$\\{(.*?)\\}")
	for keepResolving && loops < len(subMap) {
		loops++
		for k, value := range subMap {
			switch value.(type) {
			case map[string]interface{}:
				subValues(fullMap, value.(map[string]interface{}), env)
			case string:
				valStr := value.(string)
				keys := re.FindAllStringSubmatch(valStr, -1)
				for _, keyToSub := range keys {
					resolvedValue := resolveSubs(fullMap, keyToSub[1], env)
					subMap[k] = strings.Replace(valStr, "${"+keyToSub[1]+"}", resolvedValue, -1)
				}
			}
		}
	}
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
