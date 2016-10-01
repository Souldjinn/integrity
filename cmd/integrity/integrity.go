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
	for _, p := range panels {
		fmt.Printf("==== %s ====\n", p.Name)
		for _, tgt := range p.Targets {
			for _, t := range p.Tests {
				check(fmt.Sprintf("%s -> %s", tgt, t.Name), fmt.Sprintf(t.Test, tgt))
			}
		}
	}
}

func check(name string, url string) {
	client := http.DefaultClient

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s:\n%s\n", name, body)
}