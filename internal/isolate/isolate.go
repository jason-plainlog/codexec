package isolate

import (
	"codexec/internal/config"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var cfg = config.GetConfig()
var availableIDs = make(chan int, cfg.MAX_SANDBOX)

// should be called only once when the server starts
func Init() {
	for i := 0; i < cfg.MAX_SANDBOX; i++ {
		availableIDs <- i
	}
}

type (
	Limits struct {
		Time     float32 `json:"time"`
		Memory   int     `json:"memory"`
		Process  int     `json:"process"`
		FileSize int     `json:"filesize"`
		Network  bool    `json:"network"`
	}

	Sandbox struct {
		Id   string
		Path string
	}
)

func New() (Sandbox, error) {
	id := <-availableIDs
	exec.Command("isolate", "--cg", "--cleanup", "--box-id", fmt.Sprint(id)).Run()

	sandbox := Sandbox{
		Id: fmt.Sprint(id),
	}

	args := []string{
		"--box-id", sandbox.Id,
		"--cg",
		"--init",
	}

	cmd := exec.Command("isolate", args...)
	output, err := cmd.Output()
	if err != nil {
		availableIDs <- id
		return Sandbox{}, err
	}

	sandbox.Path = strings.TrimSuffix(string(output), "\n")

	return sandbox, nil
}

func (s *Sandbox) CleanUp() error {
	err := exec.Command("isolate", "--box-id", s.Id, "--cg", "--cleanup").Run()
	if err != nil {
		return err
	}

	id, _ := strconv.Atoi(s.Id)
	availableIDs <- id

	return nil
}

func (s *Sandbox) Run(command []string, limits Limits, stdin []byte) (map[string]string, error) {
	args := []string{
		"--box-id", s.Id,
		"--cg",
		"-M", s.Path + "/meta",
		"-t", fmt.Sprint(limits.Time),
		"-w", fmt.Sprint(limits.Time * 2),
		"-m", fmt.Sprint(limits.Memory),
		"-f", fmt.Sprint(limits.FileSize),
		fmt.Sprintf("--processes=%d", limits.Process),
		"-o", "stdout.txt",
		"-r", "stderr.txt",
	}
	if limits.Network {
		args = append(args, "--share-net")
	}
	args = append(args, "--run")
	args = append(args, command...)

	cmd := exec.Command("isolate", args...)
	fmt.Println(cmd.String())

	Stdin, _ := cmd.StdinPipe()
	Stdin.Write(stdin)

	err := cmd.Run()

	metaContent, _ := os.ReadFile(s.Path + "/meta")
	meta := ParseMeta(string(metaContent))

	return meta, err
}
