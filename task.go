package integrity

import (
	"time"
	"fmt"
	"github.com/robfig/cron"
)

// Task records the parameters of a test run.
type Task struct {
	Schedule string
	// resource identifier to find this task again.
	TaskName string `json:"-"`
	// defines a list of resources to target
	Targets []string
	// defines a list of tests to run by path
	Tests []struct {
		// name unifies lets you compare results across tasks and targets, even if path changes
		Name string
		// path is a URL with %s that takes a target and gets back results
		Path string
	}
}

type TaskResults struct {
	Task
	StartTime  time.Time
	FinishTime time.Time
	Results    []Result
}

// taskJob represents a diagnostic testing task that can be
// scheduled.
func TaskJob(r chan TaskResults, p Task, runner chan TestCase) cron.FuncJob {
	return func() {
		q := TaskResults{}
		q.StartTime = time.Now()

		fmt.Printf("Running %s\n", p.TaskName)

		// outer product of targets and tests.
		callback := make(chan Result)

		expected := len(p.Targets) * len(p.Tests)
		go func() {
			j := 0
			var results []Result
			for {
				j++
				i, more := <-callback
				if more {
					// Collect results
					results = append(results, i)
				}
				if !more || j >= expected {
					q.Task = p
					q.FinishTime = time.Now()
					q.Results = results
					r <- q
					// Write out results
					fmt.Printf("All done with %s:\n", p.TaskName)
					for _, k := range results {
						fmt.Printf("  [%s] %s @ %s \n    %v -> %s\n", k.TaskName, k.Name, k.RunTime.Format(time.RFC3339), k.Result, k.Note)
					}
					close(callback)
					return
				}
			}
		}()

		// sends outer product of targets and test to http queue.
		for _, tgt := range p.Targets {
			for _, t := range p.Tests {
				runner <- TestCase{
					Path:     t.Path,
					Name:     t.Name,
					Target:   tgt,
					TaskName: p.TaskName,
					Callback: callback,
				}
			}
		}
	}
}
