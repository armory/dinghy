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
	"io"
	"os"
	"os/user"

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

func CopyToLocalSpinnaker(src, dst string) (int64, error) {
	usr, err := user.Current()
	if err != nil {
		return 0, err
	}

	os.Mkdir(fmt.Sprintf("%s/.spinnaker", usr.HomeDir), 0755)

	dst = fmt.Sprintf("%s/.spinnaker/%s", usr.HomeDir, dst)

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
