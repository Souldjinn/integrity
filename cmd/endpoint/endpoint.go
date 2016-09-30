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
	"flag"
)

type Result struct {
	Result bool
	Note string
}

func main() {
	port := flag.Int("p", 3456, "port")
	flag.Parse()
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		result := Result{
			Result: true,
			Note: fmt.Sprintf("test ran on %d and %s", *port, r.URL.Path),
		}
		if v, err := json.MarshalIndent(result, "", "  "); err != nil {
			panic(err)
		} else {
			fmt.Fprint(w, string(v))
		}
	})
	fmt.Printf(fmt.Sprintf("Serving on 0.0.0.0:%d\n", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), mux))
}