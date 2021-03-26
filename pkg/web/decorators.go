package web

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

// ContextFieldFunc supplies key value pairs for decorating log messages
type ContextFieldFunc func() map[string]interface{}


// RequestContextFields returns a ContextFieldFunc that extracts Trace and User information
// from request context
func RequestContextFields(ctx context.Context) ContextFieldFunc {
	return func() map[string]interface{} {
		fields := make(map[string]interface{})
		if tc, err := extractTraceContext(ctx); err == nil {
			fields["traceId"] = tc.TraceID
			fields["spanId"] = tc.SpanID
		}
		return fields
	}
}

// RequestContextFields returns a ContextFieldFunc that extracts Trace information
// from request headers
func RequestHeaderFields(h http.Header) ContextFieldFunc {
	return func() map[string]interface{} {
		fields := make(map[string]interface{})
		tc := ExtractTraceContextHeaders(h)
		if tc.TraceID != "" {
			fields["traceId"] = tc.TraceID
		}
		return fields
	}
}

// AdditionalFields returns a ContextFieldFunc that decorates a logger with
// any fields passed in.
func AdditionalFields(fields map[string]interface{}) ContextFieldFunc {
	return func() map[string]interface{} { return fields }
}

// MuxVarsExtractor extracts URL path vars from a request to decorate a logger
func MuxVarsExtractor(r *http.Request, vars ...string) ContextFieldFunc {
	fields := map[string]interface{}{}
	requestVars := mux.Vars(r)
	for _, f := range vars {
		fields[f] = requestVars[f]
	}
	return AdditionalFields(fields)
}

func DecorateLogger(logger *logrus.Logger, cf ...ContextFieldFunc) *logrus.Entry {
	var contextLogger *logrus.Entry
	for _, c := range cf {
		if contextLogger == nil {
			contextLogger = logger.WithFields(c())
			continue
		}
		contextLogger = contextLogger.WithFields(c())
	}
	return contextLogger
}
