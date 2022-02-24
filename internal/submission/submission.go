package submission

import (
	"bytes"
	"codexec/internal/config"
	"codexec/internal/isolate"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	Submission struct {
		Language   string `json:"language"`
		SourceCode []byte `json:"source_code"`

		Tasks []Task `json:"tasks"`
		Wait  bool   `json:"wait"`
	}

	Task struct {
		Stdin  []byte         `json:"stdin"`
		Limits isolate.Limits `json:"limits"`

		Token          uuid.UUID `json:"token,omitempty"`
		CallbackURL    string    `json:"callback_url"`
		ExpectedOutput []byte    `json:"expected_output"`
	}

	Result struct {
		Stdout []byte  `json:"stdout"`
		Stderr []byte  `json:"stderr"`
		Time   float32 `json:"time"`
		Memory int     `json:"memory"`
		Status string  `json:"status"`
	}

	SubmissionResult struct {
		Results []Result `json:"results"`
	}
)

var cfg = config.GetConfig()

func (s *Submission) Check() error {
	// check tasks
	if l := len(s.Tasks); l > cfg.MAX_TASKS || l == 0 {
		return fmt.Errorf("the length of tasks should be within [1, %d]", cfg.MAX_TASKS)
	}

	for i := range s.Tasks {
		if err := s.Tasks[i].Check(); err != nil {
			return err
		}
	}

	// check language
	if _, ok := LanguageHandlers[s.Language]; !ok {
		return fmt.Errorf("invalid language")
	}

	return nil
}

func (t *Task) Check() error {
	// check for limits
	if t.Limits.Time <= 0 {
		t.Limits.Time = cfg.MAX_TIME_LIMIT
	} else if t.Limits.Time > cfg.MAX_TIME_LIMIT {
		return fmt.Errorf("the maximum time limit is %f seconds", cfg.MAX_TIME_LIMIT)
	}

	if t.Limits.Memory <= 0 {
		t.Limits.Memory = cfg.MAX_MEMORY_LIMIT
	} else if t.Limits.Memory > cfg.MAX_MEMORY_LIMIT {
		return fmt.Errorf("the maximum memory limit is %d kilobytes", cfg.MAX_MEMORY_LIMIT)
	}

	if t.Limits.FileSize <= 0 {
		t.Limits.FileSize = cfg.MAX_FILESIZE_LIMIT
	} else if t.Limits.FileSize > cfg.MAX_FILESIZE_LIMIT {
		return fmt.Errorf("the maximum filesize limit is %d kilobytes", cfg.MAX_FILESIZE_LIMIT)
	}

	if t.Limits.Process <= 0 {
		t.Limits.Process = 1
	} else if t.Limits.Process > cfg.MAX_PROCESS_LIMIT {
		return fmt.Errorf("the maximum process limit is %d", cfg.MAX_PROCESS_LIMIT)
	}

	return nil
}

func Handler(c echo.Context) error {
	submission := new(Submission)
	if err := c.Bind(submission); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := submission.Check(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if !submission.Wait {
		// return tokens immediately
		var tokens []uuid.UUID
		for i := range submission.Tasks {
			submission.Tasks[i].Token = uuid.New()
			tokens = append(tokens, submission.Tasks[i].Token)
		}

		submission.Wait = true
		body, _ := json.Marshal(submission)
		go func() {
			fmt.Println(string(body))
			resp, err := http.Post(cfg.BASE_URL+"/submission", "application/json", bytes.NewBuffer(body))
			fmt.Println(resp, err)
			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
		}()

		return c.JSON(http.StatusOK, struct {
			Tokens []uuid.UUID `json:"tokens"`
		}{
			Tokens: tokens,
		})
	} else {
		// wait for the submission result and return result by response

		for i := range submission.Tasks {
			if submission.Tasks[i].Token == uuid.Nil {
				submission.Tasks[i].Token = uuid.New()
			}
		}

		handler := LanguageHandlers[submission.Language]
		resultChan := make(chan SubmissionResult)
		go handler(submission, resultChan)

		result := <-resultChan

		return c.JSON(http.StatusOK, result)
	}
}
