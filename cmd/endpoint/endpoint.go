// Endpoint acts as a live test double for various services, so
// we don't need to run the bc app ecosystem (mainly because I
// can't right now- I'm on a plane and Athena has broken my bc
// app build... composer update took too long).
package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
)

type Result struct {
	Result bool
	Note string
}

func main() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		result := Result{
			Result: true,
			Note: "testing the tests",
		}
		if v, err := json.MarshalIndent(result, "", " "); err != nil {
			panic(err)
		} else {
			fmt.Fprint(w, string(v))
		}
	})
	log.Fatal(http.ListenAndServe("0.0.0.0:3456", mux))
}