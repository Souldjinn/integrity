// Integrity executes integrity tests against resources
// on multiple services.
//
// Theory of operation:
//  - each service exposes an endpoint to run diagnostic
//    tests.
//
//  - tests check the integrity of a resource on the
//    service. A major assumption is that each resource
//    has a singular, unique id given a url/path and
//    that ids are consistent across services.
//
package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"os"
)

// TestCase is a pipeline data structure.
type TestCase struct {
	TaskID int
	Name string
	Path string
	Target string
}

// Task records the parameters of a test run.
type Task struct {
	// resource identifier to find this task again.
	TaskID int `json:"-"`
	// defines a list of resources to target
	Targets []string
	// defines a list of tests to run by path
	Tests   []struct{
		// name unifies lets you compare results across tasks and targets, even if path changes
		Name string
		// path is a URL with %s that takes a target and gets back results
		Path string
	}
}

// Result records the result of a single test run for a task.
type Result struct {
	// taskID this result was retrieved for.
	TaskID int
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

func main() {
	fmt.Println("Act with integrity.")

	dat, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var p Task
	json.Unmarshal(dat, &p)

	// outer product of targets and tests.
	p.TaskID = 1

	in := make(chan TestCase)
	Print(Retrieve(in))

	for _, tgt := range p.Targets {
		for _, t := range p.Tests {
			in <- TestCase{
				Path: t.Path,
				Name: t.Name,
				Target: tgt,
				TaskID: p.TaskID,
			}
		}
	}
	close(in)
}

func Retrieve(in <-chan TestCase) (<-chan Result) {
	out := make(chan Result)
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
				Note string
			}
			err = json.Unmarshal(body, &r)
			if err != nil {
				panic(err)
			}

			tr := Result{}
			tr.TaskID = i.TaskID
			tr.Name = i.Name
			tr.Target = i.Target
			tr.Result = r.Result
			tr.Note = r.Note
			tr.RunTime = time.Now()
			out <- tr
		}
		close(out)
	}()
	return out
}

func Print(in <-chan Result) {
	go func() {
		for i := range in {
			fmt.Printf("  #%d %s @ %s \n    %v -> %s\n",  i.TaskID, i.Name, i.RunTime.Format(time.RFC3339), i.Result, i.Note)
		}
	}()
}