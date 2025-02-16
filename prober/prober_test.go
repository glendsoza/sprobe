package prober

import (
	"fmt"
	"net/http"
	"github.com/glendsoza/sprobe/status"
	"testing"
	"time"

	"github.com/glendsoza/sprobe/probe"
	"github.com/glendsoza/sprobe/spec"
	"github.com/stretchr/testify/assert"
)

type MockUnitsManager struct {
}

type GenericError struct {
}

func (g *GenericError) Error() string {
	return "generic"
}

func TestProberResult(t *testing.T) {
	tests := []struct {
		Status status.Status
		Output string
		Error  error
	}{

		{status.Success, "", nil},

		{status.Failure, "test", nil},
		{status.Failure, "", &GenericError{}},
	}
	for _, tt := range tests {
		expected := &ProbeResult{
			Status: tt.Status,
			Output: tt.Output,
			Error:  tt.Error,
		}
		actual := NewProbeResult().
			WithError(tt.Error).
			WithOutput(tt.Output).
			WithStatus(tt.Status)
		assert.Equal(t, expected, actual)
	}
}

type MockExecProbe struct {
	status status.Status
	output string
	err    error
}

func (me *MockExecProbe) Probe(e probe.CmdWrapper) (status.Status, string, error) {
	return me.status, me.output, me.err
}

type MockHttpProbe struct {
	status status.Status
	output string
	err    error
}

func (mh *MockHttpProbe) Probe(req *http.Request, timeout time.Duration) (status.Status, string, error) {
	return mh.status, mh.output, mh.err
}

type MockTcpProbe struct {
	status status.Status
	output string
	err    error
}

func (mt *MockTcpProbe) Probe(host string, port int, timeout time.Duration) (status.Status, string, error) {
	return mt.status, mt.output, mt.err
}

func TestProberExec(t *testing.T) {
	testSpec := &spec.LivenessProbe{
		InitialDelaySeconds: spec.ToIntRef(10),
		PeriodSeconds:       spec.ToIntRef(10),
		TimeoutSeconds:      spec.ToIntRef(10),
		FailureThreshold:    spec.ToIntRef(10),
		SuccessThreshold:    spec.ToIntRef(10),
	}
	testSpec.Exec = &spec.ExecProbe{
		Command: []string{"test"},
	}
	testCases := []struct {
		name   string
		status status.Status
		output string
		error  error
	}{
		{"normal run", status.Success, "test", nil},
		{"failed run", status.Failure, "test", fmt.Errorf("test")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			mockExecProbe := &MockExecProbe{
				status: tc.status,
				output: tc.output,
				err:    tc.error,
			}
			prober := ServiceProber{exec: mockExecProbe}
			r := prober.probe(testSpec)
			assert.Equal(tt, &ProbeResult{
				Status: tc.status,
				Output: tc.output,
				Error:  tc.error,
			}, r)
		})
	}
}

func TestProberTcp(t *testing.T) {
	testSpec := &spec.LivenessProbe{
		InitialDelaySeconds: spec.ToIntRef(10),
		PeriodSeconds:       spec.ToIntRef(10),
		TimeoutSeconds:      spec.ToIntRef(10),
		FailureThreshold:    spec.ToIntRef(10),
		SuccessThreshold:    spec.ToIntRef(10),
	}
	testSpec.TCPSocket = &spec.TCPSocketProbe{
		Port: *spec.ToIntRef(100),
	}
	testCases := []struct {
		name   string
		status status.Status
		output string
		error  error
	}{
		{"normal run", status.Success, "test", nil},
		{"failed run", status.Failure, "test", fmt.Errorf("test")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			mockTcpProbe := &MockTcpProbe{
				status: tc.status,
				output: tc.output,
				err:    tc.error,
			}
			prober := ServiceProber{tcp: mockTcpProbe}
			r := prober.probe(testSpec)
			assert.Equal(tt, &ProbeResult{
				Status: tc.status,
				Output: tc.output,
				Error:  tc.error,
			}, r)
		})
	}
}

func TestProberHttp(t *testing.T) {
	testSpec := &spec.LivenessProbe{
		InitialDelaySeconds: spec.ToIntRef(10),
		PeriodSeconds:       spec.ToIntRef(10),
		TimeoutSeconds:      spec.ToIntRef(10),
		FailureThreshold:    spec.ToIntRef(10),
		SuccessThreshold:    spec.ToIntRef(10),
	}
	testSpec.HTTPGet = &spec.HTTPGetProbe{
		Port: 100,
		Path: "http://test.com",
	}
	testCases := []struct {
		name   string
		status status.Status
		output string
		error  error
	}{
		{"normal run", status.Success, "test", nil},
		{"failed run", status.Failure, "test", fmt.Errorf("test")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			mockTcpProbe := &MockHttpProbe{
				status: tc.status,
				output: tc.output,
				err:    tc.error,
			}
			prober := ServiceProber{http: mockTcpProbe}
			r := prober.probe(testSpec)
			assert.Equal(tt, &ProbeResult{
				Status: tc.status,
				Output: tc.output,
				Error:  tc.error,
			}, r)
		})
	}
}
