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

package web

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type statusLoggingResponseWriter struct {
	http.ResponseWriter
	status    int
	bodyBytes int
}

func (w *statusLoggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *statusLoggingResponseWriter) Write(data []byte) (int, error) {
	length, err := w.ResponseWriter.Write(data)
	w.bodyBytes += length
	return length, err
}

func RequestLoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestFields := log.Fields{
			"uri":    r.URL.RequestURI(),
			"method": r.Method,
		}
		log.WithFields(requestFields).Info("incoming request")

		wrappedWriter := &statusLoggingResponseWriter{w, http.StatusOK, 0}
		defer func() {
			fields := requestFields
			fields["status"] = wrappedWriter.status
			log.WithFields(fields).Infof("outgoing request")
		}()
		h.ServeHTTP(wrappedWriter, r)
	})
}
