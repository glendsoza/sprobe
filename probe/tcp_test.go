/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http:

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"net"
	"net/http"
	"net/http/httptest"
	"github.com/glendsoza/sprobe/status"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTcpHealthChecker(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	tHost, tPortStr, err := net.SplitHostPort(server.Listener.Addr().String())
	assert.NoError(t, err)
	tPort, err := strconv.Atoi(tPortStr)
	assert.NoError(t, err)

	tests := []struct {
		host string
		port int

		expectedStatus status.Status
		expectedError  error
	}{

		{tHost, tPort, status.Success, nil},

		{tHost, -1, status.Failure, nil},
	}

	prober := NewTcpProbe()
	for _, tt := range tests {
		status, _, err := prober.Probe(tt.host, tt.port, 1*time.Second)
		assert.Equal(t, tt.expectedStatus, status)
		assert.Equal(t, tt.expectedError, err)
	}
}
