package config

import (
	"os"
)

type Config struct {
	BASE_URL    string
	MAX_SANDBOX int

	MAX_TASKS int

	MAX_TIME_LIMIT     float32
	MAX_MEMORY_LIMIT   int
	MAX_FILESIZE_LIMIT int
	MAX_PROCESS_LIMIT  int
}

func GetConfig() Config {
	config := Config{
		BASE_URL:           os.Getenv("BASE_URL"),
		MAX_SANDBOX:        1000,
		MAX_TASKS:          32,
		MAX_TIME_LIMIT:     10.0,
		MAX_MEMORY_LIMIT:   65536,
		MAX_FILESIZE_LIMIT: 1024,
		MAX_PROCESS_LIMIT:  8,
	}

	return config
}
