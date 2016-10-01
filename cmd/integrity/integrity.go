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
	"encoding/json"
	"time"
	"os"
)

type Test struct {
	Name string
	Path string
}

type TestResult struct {
	TestCase
	Result bool
	Note string
}

type TestCase struct {
	RunID int
	RunTime time.Time
	Test
	Target string
}

type Task struct {
	Targets []string
	Tests   []Test
}

func main() {
	fmt.Println("Act with integrity.")

	dat, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var p Task
	json.Unmarshal(dat, &p)

	// outer product of targets and tests.
	id := 1;

	in := make(chan TestCase)
	Print(Retrieve(in))

	for _, tgt := range p.Targets {
		for _, t := range p.Tests {
			c := TestCase{}
			c.Path = t.Path
			c.Name = t.Name
			c.Target = tgt
			c.RunID = id
			in <- c
		}
	}
}

func Retrieve(in <-chan TestCase) (<-chan TestResult) {
	out := make(chan TestResult)
	go func() {
		for i := range in {
			client := http.DefaultClient

			resp, err := client.Get(fmt.Sprintf(i.Path, i.Target))
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var r struct {
				Result bool
				Note string
			}
			err = json.Unmarshal(body, &r)
			if err != nil {
				panic(err)
			}

			tr := TestResult{TestCase: i}
			tr.Result = r.Result
			tr.Note = r.Note
			tr.RunTime = time.Now()
			out <- tr
		}
		close(out)
	}()
	return out
}

func Print(in <-chan TestResult) {
	go func() {
		for i := range in {
			fmt.Printf("  #%d %s @ %s \n    %v -> %s\n",  i.RunID, i.Name, i.RunTime.Format(time.RFC3339), i.Result, i.Note)
		}
	}()
}