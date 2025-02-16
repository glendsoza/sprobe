package probe

import (
	"net"
	"github.com/glendsoza/sprobe/status"
	"strconv"
	"time"
)

type TcpProbe interface {
	Probe(host string, port int, timeout time.Duration) (status.Status, string, error)
}

type tcpProbe struct{}

func NewTcpProbe() TcpProbe {
	return tcpProbe{}
}

func (pr tcpProbe) Probe(host string, port int, timeout time.Duration) (status.Status, string, error) {
	return DoTCPProbe(net.JoinHostPort(host, strconv.Itoa(port)), timeout)
}

func DoTCPProbe(addr string, timeout time.Duration) (status.Status, string, error) {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return status.Failure, err.Error(), nil
	}
	conn.Close()
	return status.Success, "", nil
}
