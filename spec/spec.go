package spec

import (
	"errors"
)

type ExecProbe struct {
	Command []string `yaml:"command"`
}

type HTTPGetProbe struct {
	Path        string `yaml:"path"`
	Port        int    `yaml:"port"`
	HTTPHeaders []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"httpHeaders,omitempty"`
}

type TCPSocketProbe struct {
	Port int `yaml:"port"`
}

type LivenessProbe struct {
	ServiceName         string          `yaml:"serviceName"`
	Exec                *ExecProbe      `yaml:"exec,omitempty"`
	HTTPGet             *HTTPGetProbe   `yaml:"httpGet,omitempty"`
	TCPSocket           *TCPSocketProbe `yaml:"tcpSocket,omitempty"`
	InitialDelaySeconds *int            `yaml:"initialDelaySeconds"`
	PeriodSeconds       *int            `yaml:"periodSeconds"`
	TimeoutSeconds      *int            `yaml:"timeoutSeconds"`
	FailureThreshold    *int            `yaml:"failureThreshold"`
	SuccessThreshold    *int            `yaml:"successThreshold"`
	AutoRestart         *bool           `yaml:"autoRestart"`
}

func (lp *LivenessProbe) Validate() error {
	if lp.ServiceName == "" {
		return errors.New("no service name defined; must define the service name")
	}
	definedCount := 0

	if lp.Exec != nil {
		definedCount++
	}
	if lp.HTTPGet != nil {
		definedCount++
	}
	if lp.TCPSocket != nil {
		definedCount++
	}

	if definedCount == 0 {
		return errors.New("no liveness probe type defined; must define one of exec, httpGet, or tcpSocket")
	}
	if definedCount > 1 {
		return errors.New("only one liveness probe type can be defined; multiple found")
	}

	if lp.AutoRestart == nil {
		lp.AutoRestart = ToBoolRef(false)
	}

	if lp.FailureThreshold == nil {
		lp.FailureThreshold = ToIntRef(1)
	}

	if lp.SuccessThreshold == nil {
		lp.SuccessThreshold = ToIntRef(1)
	}

	if lp.PeriodSeconds == nil {
		lp.PeriodSeconds = ToIntRef(30)
	}

	if lp.TimeoutSeconds == nil {
		lp.TimeoutSeconds = ToIntRef(10)
	}

	if lp.InitialDelaySeconds == nil {
		lp.InitialDelaySeconds = ToIntRef(10)
	}

	return nil
}

func ToBoolRef(b bool) *bool {
	return &b
}

func ToIntRef(i int) *int {
	return &i
}
