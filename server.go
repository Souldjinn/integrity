package integrity

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Integrity struct {
	TestResults map[string]TaskResults
	ResultChan chan TaskResults
}

func NewIntegrityServer() *Integrity {
	i := &Integrity{
		TestResults: make(map[string]TaskResults, 0),
		ResultChan: make(chan TaskResults),
	}
	// manage access to mutable state
	go func() {
		for z := range i.ResultChan {
			i.TestResults[z.TaskName] = z
		}
	}()
	return i
}

func (i *Integrity) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	test := r.URL.Query().Get("test")
	if test != "" {
		if r, ok := i.TestResults[test]; ok {
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
		for k := range i.TestResults {
			s = append(s, k)
		}
		m, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "%s\n", m)
	}
}