package integrity_test

import (
	"github.com/bigcommerce-labs/integrity"
	"testing"
	"time"
)

func TestTaskJob(t *testing.T) {
	task := integrity.Task{
		Schedule: "@every 10s",
		TaskName: "test task",
		Targets: []string{
			"1234",
			"4567",
		},
		Tests: []struct {
			Name string
			Path string
		}{
			{
				Name: "hello",
				Path: "http://example.org/test1/%s",
			},
			{
				Name: "hello",
				Path: "http://example.org/test2/%s",
			},
		},
	}

	// fake test runner.
	tests := make(chan integrity.TestCase)
	results := make(chan integrity.TaskResults)
	go func() {
		for t := range tests {
			t.Callback <- integrity.Result{
				TaskName: "task_name",
				Name:     "test_name",
				Target:   "target",
				RunTime:  time.Now(),
				Result:   true,
				Note:     "fake result",
			}
		}
	}()

	integrity.TaskJob(results, task, tests)()

	out := <-results

	if len(out.Results) != 4 {
		t.Logf("not enough results! got %d wanted %d", len(out.Results), 4)
		t.Fail()
	}
}
