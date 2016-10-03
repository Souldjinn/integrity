// TODO kill zombie tasks
// TODO detect leaks
// TODO serve result sets live
// TODO serve status (# routines, memory, etc)
// TODO wire in statsd
// TODO auth on production
// TODO testing and standards
// TODO deployment
package main

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	Results    []Result
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

var testresults = make(map[string]taskResults, 0)

func main() {
	fmt.Println("Act with integrity.")
	runner := make(chan TestCase)
	// number of retrieval request workers.
	for w := 0; w < 5; w++ {
		retrieve(runner)
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

		c.AddFunc(p.Schedule, taskJob(p, runner))
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
func taskJob(p Task, runner chan TestCase) cron.FuncJob {
	return func() {
		q := taskResults{}
		q.StartTime = time.Now()

		fmt.Printf("Running %s\n", p.TaskName)

		// outer product of targets and tests.
		callback := make(chan Result)

		expected := len(p.Targets) * len(p.Tests)
		go func() {
			j := 0
			var results []Result
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
				runner <- TestCase{
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

func retrieve(in <-chan TestCase) {
	go func() {
		for i := range in {
			client := http.DefaultClient

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
