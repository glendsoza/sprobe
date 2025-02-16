package probe

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/glendsoza/sprobe/status"
)

type CmdWrapper interface {
	Start() error
	SetStderr(io.Writer)
	SetStdout(io.Writer)
	Wait() error
}

type Cmd struct {
	*exec.Cmd
}

func (e *Cmd) SetStderr(w io.Writer) {
	e.Cmd.Stderr = w
}

func (e *Cmd) SetStdout(w io.Writer) {
	e.Cmd.Stdout = w
}

type ExecProbe interface {
	Probe(e CmdWrapper) (status.Status, string, error)
}

type execProbe struct{}

func NewExecProbe() ExecProbe {
	return &execProbe{}
}

func (pr *execProbe) Probe(e CmdWrapper) (status.Status, string, error) {
	writer := bytes.NewBuffer([]byte(""))
	e.SetStderr(writer)
	e.SetStdout(writer)
	err := e.Start()
	if err == nil {
		err = e.Wait()
	}
	data := writer.Bytes()

	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if ok {
			if exit.ExitCode() == 0 {
				return status.Success, string(data), nil
			}
			return status.Failure, string(data), nil
		}
		return status.Unknown, "", err
	}
	return status.Success, string(data), nil
}
