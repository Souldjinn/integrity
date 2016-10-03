// TODO
//   zombie tasks
//   routine, memory leaks
//   race conditions
//   regressions
//   edge cases (404s, test run failures, panics)
package main

import (
	"encoding/json"
	"fmt"
	"github.com/bigcommerce-labs/integrity"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var testresults map[string]integrity.TaskResults

func main() {
	fmt.Println("Act with integrity.")
	client := http.DefaultClient

	testresults = make(map[string]integrity.TaskResults, 0)
	testchan := make(chan integrity.TaskResults)

	// manage access to mutable state
	go func() {
		for z := range testchan {
			testresults[z.TaskName] = z
		}
	}()

	tests := make(chan integrity.TestCase)
	// number of retrieval request workers.
	for w := 0; w < 5; w++ {
		integrity.TestWorker(client, tests)
	}

	// add all the tasks.
	a, err := filepath.Glob(fmt.Sprintf("%s*.json", os.Args[1]))
	if err != nil {
		panic(err)
	}
	c := cron.New()
	for _, f := range a {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		var p integrity.Task
		err = json.Unmarshal(dat, &p)
		if err != nil {
			panic(err)
		}

		base := filepath.Base(f)
		p.TaskName = base[0 : len(base)-len(filepath.Ext(base))]

		c.AddFunc(p.Schedule, integrity.TaskJob(testchan, p, tests))
	}
	c.Start()

	// wait forever and let cron do its thing- may
	// be replaced with an http handler eventually.
	http.HandleFunc("/", serveHTTP)
	log.Fatal(http.ListenAndServe("0.0.0.0:4567", nil))
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	test := r.URL.Query().Get("test")
	if test != "" {
		if r, ok := testresults[test]; ok {
			// single test - show results
			m, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "%s\n", m)
		} else {
			// 404
			w.WriteHeader(404)
			fmt.Fprintf(w, "404 Not Found")
		}
	} else {
		// index page - show list of tests
		var s []string
		for k := range testresults {
			s = append(s, k)
		}
		m, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "%s\n", m)
	}
}
