package util

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// GetenvOrDefault will return the value of the given enrvironment variable,
// or, if it's blank, will return the defaultVal.
func GetenvOrDefault(envVar, defaultVal string) string {
	if val, found := os.LookupEnv(envVar); found {
		log.Infof("Checking ENV for %s...  Found: \"%s\"", envVar, val)
		return val
	}
	log.Infof("Checking ENV for %s...  Using default \"%s\"", envVar, defaultVal)
	return defaultVal
}
