package formatters

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

type testInput struct {
	hostname   string
	customerID string
	version    string
}

func TestNewHttpLogFormatter(t *testing.T) {
	cases := map[string]struct {
		input         testInput
		expectedError error
	}{
		"everything supplied, no error": {
			input:         testInput{hostname: "t", customerID: "t", version: "t"},
			expectedError: nil,
		},
		"missing argument, error": {
			input:         testInput{hostname: "t", customerID: "t"},
			expectedError: fmt.Errorf(HTTP_FORMAT_VALIDATION_ERRROR, "version"),
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			_, err := NewHttpLogFormatter(c.input.hostname, c.input.customerID, c.input.version)
			if err == nil && c.expectedError == nil {
				return
			}

			if err != nil && c.expectedError == nil {
				t.Fatalf("got an error when we didn't expect one %s", err.Error())
			}

			if err.Error() != c.expectedError.Error() {
				t.Fatalf("go unexpected error message. expected %s\ngot %s", c.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestHttpLogFormatter_Format(t *testing.T) {
	cases := map[string]struct {
		formatterInput testInput
		entry          *logrus.Entry
		expected       string
	}{
		"message is in correct format": {
			formatterInput: testInput{
				hostname:   "test1",
				customerID: "test2",
				version:    "test3",
			},
			entry: &logrus.Entry{
				Message: "hello from the test",
				Level:   logrus.InfoLevel,
			},
			expected: "test1 test2 test3 main INFO golang -- hello from the test",
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			f, _ := NewHttpLogFormatter(c.formatterInput.hostname, c.formatterInput.customerID, c.formatterInput.version)
			out, _ := f.Format(c.entry)
			if string(out) != c.expected {
				t.Fatalf("message formatted improperly. expected %s\ngot%s", c.expected, string(out))
			}
		})
	}
}
