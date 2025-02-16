package sysd

import (
	"context"
	"fmt"
	"testing"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/stretchr/testify/assert"
)

var dummyError error = fmt.Errorf("this is dummy error")

type MockSysdConn struct {
	unitStatus   []dbus.UnitStatus
	code         int
	error        error
	outputString string
}

func (msc *MockSysdConn) ListUnitsContext(context.Context) ([]dbus.UnitStatus, error) {
	return msc.unitStatus, msc.error
}

func (msc *MockSysdConn) RestartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	go func() {
		ch <- msc.outputString
	}()
	return msc.code, msc.error
}

func TestExists(t *testing.T) {
	mockConn := &MockSysdConn{}
	manager := &SysdManager{
		conn: mockConn,
	}
	mockConn.unitStatus = []dbus.UnitStatus{
		{Name: "test"},
	}
	exists, err := manager.Exists("test")
	assert.True(t, exists)
	assert.NoError(t, err)
	exists, err = manager.Exists("testing")
	assert.False(t, exists)
	assert.Error(t, err)
	mockConn.error = dummyError
	exists, err = manager.Exists("test")
	assert.False(t, exists)
	assert.Error(t, err)
}

func TestRestart(t *testing.T) {
	mockConn := &MockSysdConn{}
	manager := &SysdManager{
		conn: mockConn,
	}
	mockConn.code = 1
	mockConn.outputString = "done"
	output, err := manager.Restart("test")
	assert.Equal(t, "done", output)
	assert.NoError(t, err)
	mockConn.error = dummyError
	output, err = manager.Restart("test")
	assert.Error(t, err)
}
