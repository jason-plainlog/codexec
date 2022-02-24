package submission

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func callback(t *Task, r *Result, token uuid.UUID) {
	if t.CallbackURL == "" {
		return
	}

	client := http.Client{
		Timeout: time.Second * 10.0,
	}

	body, _ := json.Marshal(struct {
		Token  uuid.UUID `json:"token"`
		Result *Result   `json:"result"`
	}{
		Token:  token,
		Result: r,
	})

	for i := 0; i < 3; i++ {
		resp, err := client.Post(t.CallbackURL, "application/json", bytes.NewBuffer(body))
		if err == nil && resp.StatusCode == 200 {
			break
		}
	}
}
