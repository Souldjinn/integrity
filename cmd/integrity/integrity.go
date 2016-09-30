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

type Test struct {
	Name string
	Test string
}

func main() {
	fmt.Println("Act with integrity.")

	targets := []string {
		"1234",
		"5556",
	}

	tests := []Test{
		{
			Name: "6-diagnostic1",
			Test: "http://0.0.0.0:3456/%s/test/diagnostic1",
		},
		{
			Name: "7-diagnostic1",
			Test: "http://0.0.0.0:3457/%s/test/diagnostic2",
		},
	}

	// outer product of targets and tests.
	for _, tgt := range targets {
		for _, t := range tests {
			check(fmt.Sprintf("%s -> %s", tgt, t.Name), fmt.Sprintf(t.Test, tgt))
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