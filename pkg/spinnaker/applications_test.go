package spinnaker

import (
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/gorilla/mux"
	"github.com/magiconair/properties/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockServerOpts struct {
	method  string
	code    int
	payload string
}

func mockFront50(t *testing.T, opts mockServerOpts) *httptest.Server {
	r := http.NewServeMux()
	r.HandleFunc("/v2/applications", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(opts.code)
		w.Write([]byte(opts.payload))
	})
	return httptest.NewServer(r)
}

func getTestApps(t *testing.T, svr *httptest.Server) []string {
	defer svr.Close()
	settings.S.Front50.BaseURL = svr.URL
	return Applications()
}

func TestApplications(t *testing.T) {
	cases := []struct {
		opts     mockServerOpts
		expected []string
	}{
		{
			opts:     mockServerOpts{code: http.StatusOK, payload: `[{"name": "test"}, {"name": "test2"}]`},
			expected: []string{"test", "test2"},
		},
		{
			opts:     mockServerOpts{code: http.StatusBadRequest, payload: `[]`},
			expected: []string{},
		},
	}

	for _, c := range cases {
		svr := mockFront50(t, c.opts)
		apps := getTestApps(t, svr)
		assert.Equal(t, apps, c.expected)
	}
}

type mockOrcaOpts struct {
	submitCode   int
	submitResult string
	pollCode     int
	pollResult   string
}

func mockOrca(t *testing.T, opts mockOrcaOpts) *httptest.Server {
	r := mux.NewRouter()
	r.HandleFunc("/ops", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(opts.submitCode)
		w.Write([]byte(opts.submitResult))
	}).Methods(http.MethodPost)

	r.HandleFunc("/ref/{refId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(opts.pollCode)
		w.Write([]byte(opts.pollResult))
	}).Methods(http.MethodGet)

	return httptest.NewServer(r)
}

func TestNewApplication(t *testing.T) {
	cases := []struct {
		appName     string
		ownerEmail  string
		opts        mockOrcaOpts
		err         error
		front50Opts mockServerOpts
	}{
		{
			appName:    "hello",
			ownerEmail: "hello",
			opts: mockOrcaOpts{
				submitCode:   http.StatusOK,
				submitResult: `{"ref": "ref/12345"}`,
				pollCode:     http.StatusOK,
				pollResult:   `{"status": "COMPLETED", "endTime": 1}`,
			},
			front50Opts: mockServerOpts{
				code:    http.StatusOK,
				payload: `[{"name": "hello"}]`,
			},
			err: nil,
		},
	}

	for _, c := range cases {
		fakeOrca := mockOrca(t, c.opts)
		fakeFront50 := mockFront50(t, c.front50Opts)
		err := newApplicationFetch(fakeOrca, fakeFront50, c.appName, c.ownerEmail)
		assert.Equal(t, err, c.err)
	}
}

func newApplicationFetch(svr, svr2 *httptest.Server, name, email string) error {
	defer svr.Close()
	defer svr2.Close()
	settings.S.Orca.BaseURL = svr.URL
	settings.S.Front50.BaseURL = svr2.URL
	return NewApplication(name, email)
}
