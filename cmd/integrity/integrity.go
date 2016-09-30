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

	client := http.DefaultClient

	resp, err := client.Get("http://0.0.0.0:3456/1234/test/diagnostic1")
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("body: %s", 	body)
}