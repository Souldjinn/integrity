package integrity_test

import (
	"bytes"
	"github.com/bigcommerce-labs/integrity"
	"io/ioutil"
	"net/http"
	"testing"
	"fmt"
)



// Wires up a TestWorker to channels and a stubbed http client
// and inspects the results.
func TestTestWorker(t *testing.T) {
	tests := make(chan integrity.TestCase)
	results := make(chan integrity.Result)

	samples := map[string]struct{
		responseCode int
		responseBody string
		responseError error
		taskName string
		testName string
		target string
		path string
		expectedResult bool
		expectedNote string
	} {
		"run passing test": {
			responseCode: 200,
			responseBody: `{"result":true,"note":"passing test"}`,
			responseError:  nil,
			taskName: "unit",
			testName: "random_test",
			target:   "1234",
			path: "http://example.org/test/%s",
			expectedResult: true,
			expectedNote:   "passing test",
		},
		"run failing test": {
			responseCode: 200,
			responseBody: `{"result":false,"note":"failing test"}`,
			responseError:  nil,
			taskName: "unit",
			testName: "random_test",
			target:   "1234",
			path: "http://example.org/test/%s",
			expectedResult: false,
			expectedNote:   "failing test",
		},
		"response is garbled": {
			responseCode: 200,
			responseBody: `{"result":false"note":"failing test"}`,
			responseError:  nil,
			taskName: "unit",
			testName: "random_test",
			target:   "1234",
			path: "http://example.org/test/%s",
			expectedResult: false,
			expectedNote:   `invalid character '"' after object key:value pair`,
		},
		"request fails": {
			responseCode: 200,
			responseBody: `{"result":false"note":"failing test"}`,
			responseError:  fmt.Errorf("http error"),
			taskName: "unit",
			testName: "random_test",
			target:   "1234",
			path: "http://example.org/test/%s",
			expectedResult: false,
			expectedNote:   `Get http://example.org/test/1234: http error`,
		},
		"status code not OK": {
			responseCode: 500,
			responseBody: `{"result":true,"note":"passing test"}`,
			responseError:  nil,
			taskName: "unit",
			testName: "random_test",
			target:   "1234",
			path: "http://example.org/test/%s",
			expectedResult: false,
			expectedNote:   "status code: 500",
		},
	}

	for k, s := range samples {
		client := http.DefaultClient
		client.Transport = T{
			Code: s.responseCode,
			Body: s.responseBody,
			Err: s.responseError,
		}
		integrity.TestWorker(client, tests)
		tests <- integrity.TestCase{
			TaskName: s.taskName,
			Name:     s.testName,
			Path:     s.path,
			Target:   s.target,
			Callback: results,
		}
		result := <-results
		t.Logf("%+v", result)
		if result.Name != s.testName {
			t.Logf("[%s] result.Name: got %s wanted %s", k, result.Name, s.testName)
			t.Fail()
		}
		if result.TaskName != s.taskName {
			t.Logf("[%s] result.Name: got %s wanted %s", k, result.TaskName, s.taskName)
			t.Fail()
		}
		if result.Target != s.target {
			t.Logf("[%s] result.Target: got %s wanted %s", k, result.Target, s.target)
			t.Fail()
		}
		if result.Note != s.expectedNote {
			t.Logf("[%s] result.Note: got %s wanted %s", k, result.Note, s.expectedNote)
			t.Fail()
		}
		if result.Result != s.expectedResult {
			t.Logf("[%s] result.Result: got %t wanted %t", k, result.Result, s.expectedResult)
			t.Fail()
		}
	}
	close(tests)
}

type T struct {
	Code int
	Body string
	Err  error
}

func (t T) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.Err != nil {
		return nil, t.Err
	}
	buf := bytes.NewBuffer([]byte(t.Body))
	return &http.Response{
		Status:        http.StatusText(t.Code),
		StatusCode:    t.Code,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Request:       r,
		Header:        http.Header{},
		Close:         true,
		ContentLength: int64(buf.Len()),
		Body:          ioutil.NopCloser(buf),
	}, nil
}
