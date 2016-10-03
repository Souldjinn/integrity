package integrity_test

import (
	"bytes"
	"github.com/bigcommerce-labs/integrity"
	"io/ioutil"
	"net/http"
	"testing"
)

// Wires up a TestWorker to channels and a stubbed http client
// and inspects the results.
func TestTestWorker(t *testing.T) {
	tests := make(chan integrity.TestCase)
	results := make(chan integrity.Result)
	client := http.DefaultClient
	client.Transport = T{
		Code: 200,
		Body: `{"result":true,"note":"passing test"}`,
		Err:  nil,
	}
	integrity.TestWorker(client, tests)
	tests <- integrity.TestCase{
		TaskName: "unit",
		Name:     "random_test",
		Path:     "http://example.org/test/%s",
		Target:   "1234",
		Callback: results,
	}
	close(tests)
	result := <-results

	if result.Name != "random_test" {
		t.Logf("result.Name: got %s wanted %s", result.Name, "random_test")
		t.Fail()
	}
	if result.TaskName != "unit" {
		t.Logf("result.Name: got %s wanted %s", result.TaskName, "unit")
		t.Fail()
	}
	if result.Target != "1234" {
		t.Logf("result.Target: got %s wanted %s", result.Target, "1234")
		t.Fail()
	}
	if result.Note != "passing test" {
		t.Logf("result.Note: got %s wanted %s", result.Note, "passing test")
		t.Fail()
	}
	if !result.Result {
		t.Logf("result.Result: got %t wanted %t", result.Result, true)
		t.Fail()
	}
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
