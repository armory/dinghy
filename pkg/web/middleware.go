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
	"context"
	"errors"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/stdout"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
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

type TraceSettings interface {
	TraceExtract() func(handler http.Handler) http.Handler
}

type TraceContextKey struct{}

func extractTraceContext(ctx context.Context) (*TraceContext, error) {
	v, ok := ctx.Value(TraceContextKey{}).(TraceContext)
	if !ok {
		return nil, errors.New("unable to extract trace context from request")
	}
	return &v, nil
}


// TraceContext maps to the w3c traceparent header  https://w3c.github.io/trace-context/#traceparent-header
type TraceContext struct {
	Version string
	TraceID string
	// SpanID is also known as the parent id
	// since that is the upstream service's active span when the request was made
	SpanID  string
	Sampled string
}
func ExtractTraceContextHeaders(headers http.Header) TraceContext {
	var tc TraceContext
	tp := headers.Get("traceparent")
	// traceparent follows format: ${supportedVersion}-${traceID}-${spanID}-${samplingRate}
	if tp != "" {
		tpParts := strings.Split(tp, "-")
		// Only accept properly formatted header
		if len(tpParts) == 4 {
			// We only need trace id for now since version, sampling rate, and span id aren't relevant for correlating logs to traces
			tc.TraceID = tpParts[1]
		}
	} else {
		exporter, err := stdout.NewExporter(
			stdout.WithPrettyPrint(),
		)
		if err != nil {
			log.Fatalf("failed to initialize stdout export pipeline: %v", err)
		}
		ctx := context.Background()
		bsp := sdktrace.NewBatchSpanProcessor(exporter)
		tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))
		tracer := tp.Tracer("dinghyTracer")
		var parentSpan trace.Span
		ctx, parentSpan = tracer.Start(ctx, "parentSpan")
		defer parentSpan.End()
		tc.TraceID = parentSpan.SpanContext().TraceID.String()

		// Handle this error in a sensible manner where possible
		defer func() { _ = tp.Shutdown(ctx) }()

	}
	return tc
}
