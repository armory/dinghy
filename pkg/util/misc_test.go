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
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"testing"
)

const DINGHY_ENV_TEST = "DINGHY_ENV_TEST"

func TestGetenvOrDefault(t *testing.T) {
	err := os.Setenv(DINGHY_ENV_TEST, "test")
	if err != nil {
		t.Error(err)
	}

	test := GetenvOrDefault(DINGHY_ENV_TEST, "foo")
	assert.Equal(t, "test", test)

	notFound := GetenvOrDefault("DOES_NOT_EXIST", "baz")
	assert.Equal(t, "baz", notFound)
}

func TestCopyToLocalSpinnaker(t *testing.T) {
	err := CopyToLocalSpinnaker("testdata/copyfile.yml", "copyfile.yml")

	assert.Nil(t, err)
	if u, err := user.Current(); err == nil {
		if _, err := os.Stat(fmt.Sprintf("%s/.spinnaker/%s", u.HomeDir, "copyfile.yml")); err != nil {
			assert.Equal(t, false, os.IsNotExist(err))
		}
	}

}
