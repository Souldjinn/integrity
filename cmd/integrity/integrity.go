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
	Target string
}

func main() {
	fmt.Println("Act with integrity.")

	tests := []Test{
		{
			Name: "6-diagnostic1",
			Target: "http://0.0.0.0:3456/1234/test/diagnostic1",
		},
		{
			Name: "7-diagnostic1",
			Target: "http://0.0.0.0:3457/1234/test/diagnostic1",
		},
	}

	for _, t := range tests {
		check(t.Name, t.Target)
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