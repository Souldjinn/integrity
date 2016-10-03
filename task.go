package integrity

import (
	"github.com/robfig/cron"
	"time"
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

// TaskResults represent the results of running a task.
type TaskResults struct {
	Task
	StartTime  time.Time
	FinishTime time.Time
	Results    []Result
}

// TaskJob represents a diagnostic testing task that can be scheduled.
func TaskJob(r chan TaskResults, p Task, runner chan TestCase) cron.FuncJob {
	return func() {
		q := TaskResults{}
		q.StartTime = time.Now()
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
					close(callback)
					return
				}
			}
		}()
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
