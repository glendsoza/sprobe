package prober

import (
	"fmt"
	"github.com/glendsoza/sprobe/spec"
	"sync"
	"time"

	"github.com/glendsoza/sprobe/health"
	"github.com/glendsoza/sprobe/status"
	"github.com/glendsoza/sprobe/sysd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/rs/zerolog/log"
)

var (
	healthMetrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sprobe_service_health",
		Help: "Health status of services: 0 = healthy, 1 = unhealthy, -1 = unknown",
	},
		[]string{"service_name"})
)

type ServiceHealth struct {
	probeResult *ProbeResult
	health      health.Health
}

type ProberManager struct {
	prober             Prober
	serviceHealth      map[string]*ServiceHealth
	serviceHealthMutex sync.RWMutex
	probes             map[string]chan int
	probesMutex        sync.RWMutex
	unitsManager       sysd.Units
}

func NewProberManager(prober Prober) (*ProberManager, error) {
	unitsManger, err := sysd.New()
	if err != nil {
		return nil, err
	}
	return &ProberManager{
		prober:        prober,
		serviceHealth: make(map[string]*ServiceHealth),
		probes:        map[string]chan int{},
		unitsManager:  unitsManger,
	}, nil
}

func (pm *ProberManager) stopProbe(serviceName string) error {
	pm.probesMutex.Lock()
	defer pm.probesMutex.Unlock()
	c, ok := pm.probes[serviceName]
	if !ok {
		return fmt.Errorf("unable to find the probe %s", serviceName)
	}
	c <- 1
	delete(pm.probes, serviceName)
	delete(pm.serviceHealth, serviceName)
	return nil
}

func (pm *ProberManager) startProbe(spec *spec.LivenessProbe, stopChan chan int) {
	for {
		failureCount := 0
		successCount := 0
		time.Sleep(time.Duration(*spec.InitialDelaySeconds) * time.Second)
		ticker := time.NewTicker(time.Duration(*spec.PeriodSeconds) * time.Second)
	OUTER:
		for {
			select {
			case <-ticker.C:
				log.Info().Str("service_name", spec.ServiceName).Msg("probing")
				probeResult := pm.prober.probe(spec)
				log.Info().Str("service_name", spec.ServiceName).
					Str("status", probeResult.Status.String()).
					Str("output", probeResult.Output).
					Err(probeResult.Error).
					Msg("result")
				if probeResult.Status != status.Success {
					failureCount += 1
					if failureCount >= *spec.FailureThreshold {
						pm.updateServiceHealth(spec.ServiceName, health.UnHealthy, probeResult)
						ticker.Stop()
						output, err := pm.unitsManager.Restart(spec.ServiceName)
						log.Info().Str("service_name", spec.ServiceName).
							Str("output", output).
							Err(err).
							Msg("restarted")
						break OUTER
					}
				} else {
					failureCount = 0
					successCount += 1
					if successCount >= *spec.SuccessThreshold {
						successCount = 0
						if *spec.AutoRestart {
							pm.updateServiceHealth(spec.ServiceName, health.Healthy, probeResult)
						}
					}
				}
			case <-stopChan:
				return
			}
		}
	}
}

func (pm *ProberManager) updateServiceHealth(serviceName string, health health.Health, pr *ProbeResult) {
	healthMetrics.WithLabelValues(serviceName).Set(float64(health))
	pm.serviceHealthMutex.Lock()
	defer pm.serviceHealthMutex.Unlock()
	pm.serviceHealth[serviceName].health = health
	pm.serviceHealth[serviceName].probeResult = pr
}

func (pm *ProberManager) getServiceHealth(serviceName string) ServiceHealth {
	pm.serviceHealthMutex.RLock()
	defer pm.serviceHealthMutex.RUnlock()
	return *pm.serviceHealth[serviceName]
}

func (pm *ProberManager) Add(spec *spec.LivenessProbe) error {
	exits, err := pm.unitsManager.Exists(spec.ServiceName)
	if !exits {
		return fmt.Errorf("cannot find service %s because %s", spec.ServiceName, err)
	}
	err = spec.Validate()
	if err != nil {
		return err
	}
	pm.serviceHealthMutex.Lock()
	pm.probesMutex.Lock()
	defer pm.serviceHealthMutex.Unlock()
	defer pm.probesMutex.Unlock()
	pm.serviceHealth[spec.ServiceName] = &ServiceHealth{health: health.Unknown}
	stopChan := make(chan int)
	pm.probes[spec.ServiceName] = stopChan
	go pm.startProbe(spec, stopChan)
	return nil
}
