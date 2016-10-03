package integrity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// TestCase is a pipeline data structure.
type TestCase struct {
	TaskName string
	Name     string
	Path     string
	Target   string
	Callback chan Result
}

// Result records the result of a single test run for a task.
type Result struct {
	// TaskName this result was retrieved for.
	TaskName string
	// test name - the task can be used to recover the path
	Name string
	// target name - used with path, can reconstruct URL.
	Target string
	// time when result retrieved.
	RunTime time.Time
	// result value
	Result bool
	// human-readable explanation.
	Note string
}

// TestWorker reads tests out of a channel and runs them by
// making an HTTP request to the appropriate test endpoint.
// The results are sent back along a channel provided by the
// test case to the original Task.
func TestWorker(client *http.Client, tests <-chan TestCase) {
	go func() {
		for i := range tests {
			resp, err := client.Get(fmt.Sprintf(i.Path, i.Target))
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var r struct {
				Result bool
				Note   string
			}
			err = json.Unmarshal(body, &r)
			if err != nil {
				panic(err)
			}

			tr := Result{}
			tr.TaskName = i.TaskName
			tr.Name = i.Name
			tr.Target = i.Target
			tr.Result = r.Result
			tr.Note = r.Note
			tr.RunTime = time.Now()
			i.Callback <- tr
		}
	}()
}
