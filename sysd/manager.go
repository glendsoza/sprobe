package sysd

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

type SysdConn interface {
	ListUnitsContext(context.Context) ([]dbus.UnitStatus, error)
	RestartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error)
}

type Units interface {
	Exists(serviceName string) (bool, error)
	Restart(serviceName string) (string, error)
}

type SysdManager struct {
	conn SysdConn
}

func New() (*SysdManager, error) {
	conn, err := dbus.NewWithContext(context.Background())
	if err != nil {
		return nil, err
	}
	return &SysdManager{conn: conn}, nil
}

func (s *SysdManager) Exists(serviceName string) (bool, error) {
	units, err := s.conn.ListUnitsContext(context.Background())
	if err != nil {
		return false, err
	}
	for _, u := range units {
		if u.Name == serviceName {
			return true, nil
		}
	}
	return false, fmt.Errorf("unable to find unit with name %s", serviceName)
}

func (s *SysdManager) Restart(serviceName string) (string, error) {
	outputChan := make(chan string)
	_, err := s.conn.RestartUnitContext(context.Background(), serviceName, "replace", outputChan)
	if err != nil {
		return "", err
	}
	return <-outputChan, nil
}
