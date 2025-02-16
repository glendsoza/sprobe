/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"fmt"
	"io"
	"github.com/glendsoza/sprobe/status"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeCmd struct {
	out    []byte
	err    error
	writer io.Writer
}

func (f *FakeCmd) SetStdout(out io.Writer) {
	f.writer = out
}

func (f *FakeCmd) SetStderr(out io.Writer) {
	f.writer = out
}

func (f *FakeCmd) Start() error {
	if f.writer != nil {
		f.writer.Write(f.out)
		return f.err
	}
	return f.err
}

func (f *FakeCmd) Wait() error { return nil }

func TestExec(t *testing.T) {
	prober := NewExecProbe()

	tenKilobyte := strings.Repeat("logs-123", 128*10)      // 8*128*10=10240 = 10KB of text.
	elevenKilobyte := strings.Repeat("logs-123", 8*128*11) // 8*128*11=11264 = 11KB of text.

	tests := []struct {
		expectedStatus status.Status
		expectError    bool
		input          string
		output         string
		err            error
	}{
		{status.Success, false, "OK", "OK", nil},
		{status.Success, false, elevenKilobyte, tenKilobyte, nil},
		{status.Unknown, true, "", "", fmt.Errorf("test error")},
	}

	for _, test := range tests {
		fake := &FakeCmd{
			out: []byte(test.output),
			err: test.err,
		}
		status, output, err := prober.Probe(fake)
		assert.Equal(t, test.expectedStatus, status)
		if err != nil {
			assert.Equal(t, test.err, err)
		}
		if err == nil {
			assert.False(t, test.expectError)
		}
		if test.output != output {
			assert.Equal(t, test.output, output)
		}
	}
}
