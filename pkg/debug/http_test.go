package debug

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockDebugServer() *httptest.Server {
	r := http.NewServeMux()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(r)
}

func TestNewInterceptorHttpClient(t *testing.T) {
	m := mockDebugServer()
	defer m.Close()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	client := NewInterceptorHttpClient(logger, true)
	resp, _ := client.Get(m.URL)
	t.Run("test_new_http_client", func(t *testing.T) {
		assert.Equal(t, 200, resp.StatusCode)
	})
}
