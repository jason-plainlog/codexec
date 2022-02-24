package submission

import (
	"codexec/internal/isolate"
	"fmt"
	"os"
	"sync"
)

func python3TaskHandler(s *Submission, t *Task, r chan Result, wg *sync.WaitGroup) {
	fmt.Println("task", t)
	result := Result{}

	sandbox, err := isolate.New()
	if err != nil {
		result.Status = "Internal Error"

		callback(t, &result, t.Token)
		r <- result
		wg.Done()
		return
	}
	defer sandbox.CleanUp()

	// put source file
	os.WriteFile(sandbox.Path+"/box/source.py", s.SourceCode, 0644)

	// execute source.py and write input
	meta, _ := sandbox.Run(
		[]string{"/usr/bin/python3", "source.py"},
		t.Limits,
		t.Stdin,
	)

	result.Stdout, _ = os.ReadFile(sandbox.Path + "/box/stdout.txt")
	result.Stderr, _ = os.ReadFile(sandbox.Path + "/box/stderr.txt")

	fmt.Sscanf(meta["time"], "%f", &result.Time)
	fmt.Sscanf(meta["max-rss"], "%d", &result.Memory)
	result.Status = meta["status"]

	fmt.Println("result", result)

	go callback(t, &result, t.Token)
	r <- result

	wg.Done()
}

func python3Handler(s *Submission, r chan SubmissionResult) {
	result := SubmissionResult{Results: make([]Result, len(s.Tasks))}
	taskResultChan := make([]chan Result, len(s.Tasks))

	wg := new(sync.WaitGroup)
	wg.Add(len(s.Tasks))

	for i, task := range s.Tasks {
		// run tasks in parallel
		taskResultChan[i] = make(chan Result, 1)
		go python3TaskHandler(s, &task, taskResultChan[i], wg)
	}

	wg.Wait()
	for i := range s.Tasks {
		result.Results[i] = <-taskResultChan[i]
	}

	fmt.Println(result)

	r <- result
}
