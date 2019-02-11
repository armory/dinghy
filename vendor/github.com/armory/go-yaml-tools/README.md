[![Coverage Status](https://coveralls.io/repos/github/armory/go-yaml-tools/badge.svg?branch=master)](https://coveralls.io/github/armory/go-yaml-tools?branch=master)

## Go implementation of this library:

https://github.com/armory-io/yaml-tools/blob/master/yamltools/resolver.py


## How To Use This Package

The main package is the `spring` package.

The only exposed function is:

```
func LoadProperties(propNames []string, configDir string, envKeyPairs []string) (map[string]interface{}, error)
```


For example if you want to load the following files:

* `spinnaker.yml`
* `spinnaker-local.yml`
* `gate-local.yml`
* `gate-armory.yml`

For `propNames` you would give

```
["spinnaker", "gate"]
```

and you'll need to make sure your `envKeyPairs` has the following key pair as one of it's
variables
```
SPRING_PROFILES_ACTIVE="armory,local"
```

The `configDir` is where the configuration files live, typically `/opt/spinnaker/config` for Spinnaker files.
