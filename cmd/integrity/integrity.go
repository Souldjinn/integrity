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

func main() {
	fmt.Println("Act with integrity.")
	client := http.DefaultClient

	intg := integrity.NewIntegrityServer()

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

		c.AddFunc(p.Schedule, integrity.TaskJob(intg.ResultChan, p, tests))
	}
	c.Start()

	// wait forever and let cron do its thing- may
	// be replaced with an http handler eventually.
	log.Fatal(http.ListenAndServe("0.0.0.0:4567", intg))
}