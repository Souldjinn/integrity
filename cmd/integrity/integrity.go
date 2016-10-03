// TODO
//   zombie tasks
//   routine, memory leaks
//   race conditions
//   regressions
//   edge cases (404s, test run failures, panics)
package main

import (
	"encoding/json"
	"fmt"
	"github.com/bigcommerce-labs/integrity"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Task records the parameters of a test run.
type Task struct {
	Schedule string
	// resource identifier to find this task again.
	TaskName string `json:"-"`
	// defines a list of resources to target
	Targets []string
	// defines a list of tests to run by path
	Tests []struct {
		// name unifies lets you compare results across tasks and targets, even if path changes
		Name string
		// path is a URL with %s that takes a target and gets back results
		Path string
	}
}

type taskResults struct {
	Task
	StartTime  time.Time
	FinishTime time.Time
	Results    []integrity.Result
}

var testresults = make(map[string]taskResults, 0)

func main() {
	fmt.Println("Act with integrity.")
	client := http.DefaultClient
	tests := make(chan integrity.TestCase)
	// number of retrieval request workers.
	for w := 0; w < 5; w++ {
		integrity.TestWorker(client, tests)
	}

	// add all the tasks.
	a, err := filepath.Glob(fmt.Sprintf("%s*.json", os.Args[1]))
	if err != nil {
		panic(err)
	}
	c := cron.New()
	for _, f := range a {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		var p Task
		err = json.Unmarshal(dat, &p)
		if err != nil {
			panic(err)
		}

		base := filepath.Base(f)
		p.TaskName = base[0 : len(base)-len(filepath.Ext(base))]

		c.AddFunc(p.Schedule, taskJob(p, tests))
	}
	c.Start()

	// wait forever and let cron do its thing- may
	// be replaced with an http handler eventually.
	http.HandleFunc("/", serveHTTP)
	log.Fatal(http.ListenAndServe("0.0.0.0:4567", nil))
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	test := r.URL.Query().Get("test")
	if test != "" {
		if r, ok := testresults[test]; ok {
			// single test - show results
			m, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "%s\n", m)
		} else {
			// 404
			w.WriteHeader(404)
			fmt.Fprintf(w, "404 Not Found")
		}
	} else {
		// index page - show list of tests
		var s []string
		for k := range testresults {
			s = append(s, k)
		}
		m, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "%s\n", m)
	}
}

// taskJob represents a diagnostic testing task that can be
// scheduled.
func taskJob(p Task, runner chan integrity.TestCase) cron.FuncJob {
	return func() {
		q := taskResults{}
		q.StartTime = time.Now()

		fmt.Printf("Running %s\n", p.TaskName)

		// outer product of targets and tests.
		callback := make(chan integrity.Result)

		expected := len(p.Targets) * len(p.Tests)
		go func() {
			j := 0
			var results []integrity.Result
			for {
				j++
				i, more := <-callback
				if more {
					// Collect results
					results = append(results, i)
				}
				if !more || j >= expected {
					q.Task = p
					q.FinishTime = time.Now()
					q.Results = results
					testresults[p.TaskName] = q
					// Write out results
					fmt.Printf("All done with %s:\n", p.TaskName)
					for _, k := range results {
						fmt.Printf("  [%s] %s @ %s \n    %v -> %s\n", k.TaskName, k.Name, k.RunTime.Format(time.RFC3339), k.Result, k.Note)
					}
					close(callback)
					return
				}
			}
		}()

		// sends outer product of targets and test to http queue.
		for _, tgt := range p.Targets {
			for _, t := range p.Tests {
				runner <- integrity.TestCase{
					Path:     t.Path,
					Name:     t.Name,
					Target:   tgt,
					TaskName: p.TaskName,
					Callback: callback,
				}
			}
		}
	}
}
