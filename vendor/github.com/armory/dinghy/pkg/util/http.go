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
	"net/http"

	log "github.com/sirupsen/logrus"
)

// WriteHTTPError writes an HTTP error to an http.ResponseWriter
func WriteHTTPError(w http.ResponseWriter, code int, err error) {
	log.Error(err)
	formattedError := fmt.Errorf(`{"status": %d, "error": "%s"}`, code, err.Error())
	http.Error(w, formattedError.Error(), code)
}
