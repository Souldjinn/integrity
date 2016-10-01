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
)

type Panel struct {
	Name string
	Targets []string
	Tests []Test
}

type Test struct {
	Name string
	Test string
}

type JsonTestResult struct {
	Result bool
	Note string
}

type TestResult struct {
	TestCase
	JsonTestResult
}

type TestCase struct {
	RunID int
	RunTime time.Time
	Test
	Target string
}

func main() {
	fmt.Println("Act with integrity.")

	panels := []Panel{
		Panel{
			Name: "Distributed Diagnostics",
			Targets: []string{
				"1234",
				"5556",
			},
			Tests: []Test{
				{
					Name: "6-diagnostic1",
					Test: "http://0.0.0.0:3456/%s/test/diagnostic1",
				},
				{
					Name: "7-diagnostic1",
					Test: "http://0.0.0.0:3457/%s/test/diagnostic2",
				},
			},
		},
		Panel{
			Name: "Secondary Suite",
			Targets: []string{
				"1234",
				"5556",
				"7891",
				"4444",
			},
			Tests: []Test{
				{
					Name: "6-diagnostic2",
					Test: "http://0.0.0.0:3456/%s/test/xyz",
				},
				{
					Name: "7-diagnostic2",
					Test: "http://0.0.0.0:3457/%s/test/abc",
				},
			},
		},
	}

	// outer product of targets and tests.
	id := 1;

	in := make(chan TestCase)
	Print(Retrieve(in))

	for _, p := range panels {
		for _, tgt := range p.Targets {
			for _, t := range p.Tests {
				c := TestCase{}
				c.Test.Test = t.Test
				c.Name = t.Name
				c.Target = tgt
				c.RunID = id
				in <- c
			}
		}
		id++
	}
}

func Retrieve(in <-chan TestCase) (<-chan TestResult) {
	out := make(chan TestResult)
	go func() {
		for i := range in {
			client := http.DefaultClient

			resp, err := client.Get(fmt.Sprintf(i.Test.Test, i.Target))
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var r JsonTestResult
			err = json.Unmarshal(body, &r)
			if err != nil {
				panic(err)
			}

			tr := TestResult{i, r}
			tr.RunTime = time.Now()
			out <- tr
		}
		close(out)
	}()
	return out
}

func Print(in <-chan TestResult) {
	go func() {
		for i := range in {
			fmt.Printf("  #%d %s @ %s \n    %v -> %s\n",  i.RunID, i.Name, i.RunTime.Format(time.RFC3339), i.Result, i.Note)
		}
	}()
}