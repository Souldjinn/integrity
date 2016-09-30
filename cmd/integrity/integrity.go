// Integrity executes integrity tests against endpoints on
// services.

package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

func main() {
	fmt.Println("Act with integrity.")
	check("http://0.0.0.0:3456/1234/test/diagnostic1")
	check("http://0.0.0.0:3457/1234/test/diagnostic1")
}

func check(url string) {
	client := http.DefaultClient

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s:\n%s\n", url, body)
}