package prober

import (
	"fmt"
	"testing"
	"time"

	"github.com/glendsoza/sprobe/health"
	"github.com/glendsoza/sprobe/spec"
	"github.com/glendsoza/sprobe/status"

	"github.com/stretchr/testify/assert"
)

type DummyUnits struct {
}

func (du *DummyUnits) Exists(serviceName string) (bool, error) {
	return true, nil
}
func (du *DummyUnits) Restart(serviceName string) (string, error) {
	return "done", nil
}

var dummyTestSpec = &spec.LivenessProbe{
	ServiceName:         "test",
	InitialDelaySeconds: spec.ToIntRef(0),
	PeriodSeconds:       spec.ToIntRef(1),
	TimeoutSeconds:      spec.ToIntRef(10),
	FailureThreshold:    spec.ToIntRef(0),
	SuccessThreshold:    spec.ToIntRef(0),
}

func TestProberManager_StartStopProbe(t *testing.T) {
	mockerProber := &ServiceProber{
		exec: &MockExecProbe{status: status.Success, output: "wow", err: nil},
		http: &MockHttpProbe{},
		tcp:  &MockTcpProbe{},
	}
	pm := &ProberManager{
		prober:        mockerProber,
		unitsManager:  &DummyUnits{},
		serviceHealth: map[string]*ServiceHealth{},
		probes:        map[string]chan int{},
	}
	dummyTestSpec.Exec = &spec.ExecProbe{
		Command: []string{"test"},
	}
	pm.Add(dummyTestSpec)
	time.Sleep(2 * time.Second)
	err := pm.stopProbe(dummyTestSpec.ServiceName)
	assert.NoError(t, err)
}

func TestProberManager_ServiceHealthHealthy(t *testing.T) {
	mockerProber := &ServiceProber{
		exec: &MockExecProbe{status: status.Success, output: "wow", err: nil},
		http: &MockHttpProbe{},
		tcp:  &MockTcpProbe{},
	}
	pm := &ProberManager{
		prober:        mockerProber,
		unitsManager:  &DummyUnits{},
		serviceHealth: map[string]*ServiceHealth{},
		probes:        map[string]chan int{},
	}
	dummyTestSpec.InitialDelaySeconds = spec.ToIntRef(2)
	dummyTestSpec.Exec = &spec.ExecProbe{
		Command: []string{"test"},
	}
	pm.Add(dummyTestSpec)
	serviceHealth := pm.getServiceHealth(dummyTestSpec.ServiceName)
	assert.Equal(t, health.Unknown, serviceHealth.health)
	time.Sleep(5 * time.Second)
	serviceHealth = pm.getServiceHealth(dummyTestSpec.ServiceName)
	assert.Equal(t, ServiceHealth{probeResult: &ProbeResult{Status: status.Success, Output: "wow", Error: nil}, health: health.Healthy}, serviceHealth)
}

func TestProberManager_ServiceHealthUnHealthy(t *testing.T) {
	dummyErr := fmt.Errorf("failed")
	mockerProber := &ServiceProber{
		exec: &MockExecProbe{status: status.Failure, output: "wow", err: dummyErr},
		http: &MockHttpProbe{},
		tcp:  &MockTcpProbe{},
	}
	pm := &ProberManager{
		prober:        mockerProber,
		unitsManager:  &DummyUnits{},
		serviceHealth: map[string]*ServiceHealth{},
		probes:        map[string]chan int{},
	}
	dummyTestSpec.InitialDelaySeconds = spec.ToIntRef(2)
	dummyTestSpec.Exec = &spec.ExecProbe{
		Command: []string{"test"},
	}
	pm.Add(dummyTestSpec)
	serviceHealth := pm.getServiceHealth(dummyTestSpec.ServiceName)
	assert.Equal(t, health.Unknown, serviceHealth.health)
	time.Sleep(5 * time.Second)
	serviceHealth = pm.getServiceHealth(dummyTestSpec.ServiceName)
	assert.Equal(t, ServiceHealth{probeResult: &ProbeResult{Status: status.Failure, Output: "wow", Error: dummyErr}, health: health.UnHealthy}, serviceHealth)
}
