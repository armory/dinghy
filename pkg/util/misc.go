/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package util

import (
	"fmt"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"os"
	"os/user"
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

// GetenvOrDefaultRedact will return the value of the given enrvironment
// variable, or, if it's blank, will return the defaultVal.  Will redact any
// value found in the log output
func GetenvOrDefaultRedact(envVar, defaultVal string) string {
	if val, found := os.LookupEnv(envVar); found {
		redacted := val
		if redacted != "" {
			redacted = "**REDACTED**"
		}
		log.Infof("Checking ENV for %s...  Found: \"%s\"", envVar, redacted)
		return val
	}
	redacted := defaultVal
	if redacted != "" {
		redacted = "**REDACTED**"
	}
	log.Infof("Checking ENV for %s...  Using default \"%s\"", envVar, redacted)
	return defaultVal
}

func CopyToLocalSpinnaker(src, dst string) error {
	if u, err := user.Current(); err != nil {
		return err
	} else {
		return copy.Copy(src, fmt.Sprintf("%s/.spinnaker/%s", u.HomeDir, dst))
	}
}
