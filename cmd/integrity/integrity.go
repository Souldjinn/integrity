// Integrity executes integrity tests against endpoints on
// services.

package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

type Test struct {
	Name string
	Host string
	Test string
}

type Target struct {
	Name string
	Host map[string]string
}

func main() {
	fmt.Println("Act with integrity.")

	targets := []Target {
		{
			Name: "tgt-1234",
			Host: map[string]string{
				"3456": "http://0.0.0.0:3456/1234/test/",
				"3457": "http://0.0.0.0:3457/1234/test/",
			},
		},
		{
			Name: "tgt-5556",
			Host: map[string]string{
				"3456": "http://0.0.0.0:3456/5556/test/",
				"3457": "http://0.0.0.0:3457/5556/test/",
			},
		},
	}

	tests := []Test{
		{
			Name: "6-diagnostic1",
			Host: "3456",
			Test: "xyz",
		},
		{
			Name: "7-diagnostic1",
			Host: "3457",
			Test: "abc",
		},
	}

	// outer product of targets and tests.
	for _, tgt := range targets {
		for _, t := range tests {
			check(fmt.Sprintf("%s -> %s", tgt.Name, t.Name), fmt.Sprintf("%s/%s", tgt.Host[t.Host], t.Test))
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