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

type TestResult struct {
	Result bool
	Note string
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
	for _, p := range panels {
		fmt.Printf(">> %s (Run #%d @ %s)\n", p.Name, id, time.Now().Format(time.RFC3339))
		for _, tgt := range p.Targets {
			for _, t := range p.Tests {
				check(id, fmt.Sprintf("%s -> %s", tgt, t.Name), fmt.Sprintf(t.Test, tgt))
			}
		}
		id++
	}
}

func check(runID int, name string, url string) {
	client := http.DefaultClient

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var r TestResult
	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}

	fmt.Printf("  #%d %s @ %s \n    %v -> %s\n",  runID, name, time.Now().Format(time.RFC3339), r.Result, r.Note)
}