package spinnaker

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestDefaultFront50API_Applications(t *testing.T) {
	cases := map[string]struct {
		opts     mockServerOpts
		expected []string
	}{
		"happy path": {
			opts:     mockServerOpts{code: http.StatusOK, payload: `[{"name": "test"}, {"name": "test2"}]`},
			expected: []string{"test", "test2"},
		},
		"no applications": {
			opts:     mockServerOpts{code: http.StatusBadRequest, payload: `[]`},
			expected: []string{},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			svr := mockFront50(t, c.opts)
			defer svr.Close()
			front50 := DefaultFront50API{
				BaseURL:   svr.URL,
				APIClient: &DefaultAPIClient{},
			}
			apps := front50.Applications()
			assert.Equal(t, apps, c.expected)
		})

	}
}

type fakeOrca struct {
	OrcaAPI
	submitResponse TaskRefResponse
	submitError    error
	pollResopnse   ExecutionResponse
	pollError      error
}

func (fo *fakeOrca) SubmitTask(task Task) (*TaskRefResponse, error) {
	return &fo.submitResponse, fo.submitError
}

func (fo *fakeOrca) PollTaskStatus(refUrl string, timeout time.Duration) (*ExecutionResponse, error) {
	return &fo.pollResopnse, fo.pollError
}

func TestDefaultFront50API_NewApplication(t *testing.T) {
	cases := map[string]struct {
		err         error
		front50Opts mockServerOpts
		appSpec     ApplicationSpec
		orcaMock    *fakeOrca
	}{
		"happy path": {
			appSpec: ApplicationSpec{
				Name:  "application",
				Email: "email",
			},
			front50Opts: mockServerOpts{
				code:    http.StatusOK,
				payload: `[{"name": "application"}]`,
			},
			err: nil,
			orcaMock: &fakeOrca{
				submitResponse: TaskRefResponse{Ref: "ref/12345"},
				submitError:    nil,
				pollResopnse:   ExecutionResponse{Status: "COMPLETED", EndTime: 1},
				pollError:      nil,
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			fakeFront50 := mockFront50(t, c.front50Opts)
			front50 := DefaultFront50API{
				BaseURL:   fakeFront50.URL,
				OrcaAPI:   c.orcaMock,
				APIClient: &DefaultAPIClient{},
			}
			err := front50.NewApplication(c.appSpec)
			assert.Nil(t, err)
		})

	}
}
