package hooks

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serverMock(calls *int) (*httptest.Server, func()) {
	r := http.NewServeMux()
	r.HandleFunc("/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		*calls++
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(r)

	return ts, func() {
		ts.Close()
	}
}

type mockFormatter struct{}

func (mf mockFormatter) Format(e *logrus.Entry) ([]byte, error) {
	return nil, nil
}

func TestHttpDebugHook_Fire(t *testing.T) {
	cases := map[string]struct {
		entry         *logrus.Entry
		expectedError error
	}{
		"posts to server": {
			entry: &logrus.Entry{},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			var calls int
			svr, closeFunc := serverMock(&calls)
			defer closeFunc()
			hook := HttpDebugHook{
				Endpoint:  svr.URL + "/v1/logs",
				Formatter: mockFormatter{},
			}

			err := hook.Fire(c.entry)
			if err != nil && c.expectedError == nil {
				t.Fatalf("got an error when we didn't expect one %s", err.Error())
			}
			if calls != 1 {
				t.Fatalf("failed to actually call the server")
			}
		})
	}
}
