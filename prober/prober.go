package prober

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"github.com/glendsoza/sprobe/spec"
	"time"

	"github.com/glendsoza/sprobe/probe"
	"github.com/glendsoza/sprobe/status"
)

type ProbeResult struct {
	Status status.Status
	Output string
	Error  error
}

func NewProbeResult() *ProbeResult {
	return &ProbeResult{
		Status: status.Unknown,
		Output: "",
		Error:  nil,
	}
}

func (pr *ProbeResult) WithStatus(status status.Status) *ProbeResult {
	pr.Status = status
	return pr
}

func (pr *ProbeResult) WithOutput(output string) *ProbeResult {
	pr.Output = output
	return pr
}

func (pr *ProbeResult) WithError(err error) *ProbeResult {
	pr.Error = err
	return pr
}

type Prober interface {
	probe(spec *spec.LivenessProbe) *ProbeResult
}

type ServiceProber struct {
	exec probe.ExecProbe
	http probe.HttpProbe
	tcp  probe.TcpProbe
}

func NewServiceProber() Prober {
	return &ServiceProber{
		exec: probe.NewExecProbe(),
		http: probe.NewHttpProbe(true),
		tcp:  probe.NewTcpProbe()}
}

func (p *ServiceProber) probe(spec *spec.LivenessProbe) *ProbeResult {
	timeOutDuration := time.Duration(*spec.TimeoutSeconds) * time.Second
	switch {
	case spec.Exec != nil:
		var cmd *exec.Cmd
		ctx, cancel := context.WithTimeout(context.Background(), timeOutDuration)
		defer cancel()
		if len(spec.Exec.Command) == 1 {
			cmd = exec.CommandContext(ctx, spec.Exec.Command[0])
		} else {
			cmd = exec.CommandContext(ctx, spec.Exec.Command[0], spec.Exec.Command[1:]...)
		}

		probeStatus, output, err := p.exec.Probe(&probe.Cmd{Cmd: cmd})
		return NewProbeResult().
			WithStatus(probeStatus).
			WithOutput(output).
			WithError(err)

	case spec.HTTPGet != nil:
		req, err := http.NewRequest("GET", fmt.Sprintf("%s:%d", spec.HTTPGet.Path, spec.HTTPGet.Port), nil)
		if err != nil {
			return NewProbeResult().
				WithStatus(status.Unknown).
				WithOutput("").
				WithError(err)
		}
		for _, header := range spec.HTTPGet.HTTPHeaders {
			req.Header.Set(header.Name, header.Value)
		}
		probeStatus, output, err := p.http.Probe(req, timeOutDuration)
		return NewProbeResult().
			WithStatus(probeStatus).
			WithOutput(output).
			WithError(err)

	case spec.TCPSocket != nil:
		probeStatus, output, err := p.tcp.Probe("localhost", spec.TCPSocket.Port, timeOutDuration)
		return NewProbeResult().
			WithStatus(probeStatus).
			WithOutput(output).
			WithError(err)
	}
	return NewProbeResult().
		WithStatus(status.Unknown).
		WithOutput("").
		WithError(fmt.Errorf("unable to determine the prober from the spec"))
}
