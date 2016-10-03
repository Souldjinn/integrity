package integrity_test

import (
	"testing"
	"github.com/bigcommerce-labs/integrity"
	"time"
	"fmt"
)

func TestTaskJob(t *testing.T) {
	intg := integrity.NewIntegrityServer()
	task := integrity.Task{
		Schedule: "@every 10s",
		TaskName: "test task",
		Targets: []string{
			"1234",
			"4567",
		},
		Tests: []struct{
			Name string
			Path string
		}{
			{
				Name: "hello",
				Path: "http://example.org/test/%s",
			},
		},
	}

	// fake test runner.
	tests := make(chan integrity.TestCase)
	go func() {
		for t := range tests {
			t.Callback <- integrity.Result{
				TaskName: "task_name",
				Name: "test_name",
				Target: "target",
				RunTime: time.Now(),
				Result: true,
				Note: "fake result",
			}
		}
	}()

	integrity.TaskJob(intg, task, tests)()

	// wait for tasks to be processed
	time.Sleep(1*time.Second)

	fmt.Printf("%+v\n", intg.TestResults)
}