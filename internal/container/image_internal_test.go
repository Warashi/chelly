/*
Copyright © 2026 Shinnosuke Sawada-Dazai <3600530+Warashi@users.noreply.github.com>

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

package container

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"testing"
)

var errDaemonUnreachable = errors.New("cannot connect to the daemon")

// exitError fakes a runtime that exited non-zero.
func exitError(t *testing.T) error {
	t.Helper()

	err := exec.CommandContext(t.Context(), "false").Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("running false: %v", err)
	}

	return fmt.Errorf("running false: %w", err)
}

func TestCheckImageWith_ListRuntimes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		stdout  string
		err     error
		wantErr error
	}{
		"found":                {stdout: "0123abcd\n", err: nil, wantErr: nil},
		"missing":              {stdout: "", err: nil, wantErr: ErrImageNotFound},
		"missing with newline": {stdout: "\n", err: nil, wantErr: ErrImageNotFound},
		"check failed":         {stdout: "", err: errDaemonUnreachable, wantErr: errDaemonUnreachable},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var gotArgs []string

			output := func(_ context.Context, cmd string, args ...string) ([]byte, error) {
				gotArgs = append([]string{cmd}, args...)

				return []byte(testCase.stdout), testCase.err
			}

			err := checkImageWith(t.Context(), output, "/usr/bin/podman")
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("checkImageWith = %v, want %v", err, testCase.wantErr)
			}

			wantArgs := []string{"/usr/bin/podman", "image", "ls", "-q", ImageName}
			if !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Errorf("args: got %v, want %v", gotArgs, wantArgs)
			}
		})
	}
}

func TestCheckImageWith_AppleContainer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err     error
		wantErr error
	}{
		"found":                 {err: nil, wantErr: nil},
		"inspect exited":        {err: exitError(t), wantErr: ErrImageNotFound},
		"runtime not startable": {err: errDaemonUnreachable, wantErr: errDaemonUnreachable},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var gotArgs []string

			output := func(_ context.Context, cmd string, args ...string) ([]byte, error) {
				gotArgs = append([]string{cmd}, args...)

				return nil, testCase.err
			}

			err := checkImageWith(t.Context(), output, "container")
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("checkImageWith = %v, want %v", err, testCase.wantErr)
			}

			wantArgs := []string{"container", "image", "inspect", ImageName}
			if !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Errorf("args: got %v, want %v", gotArgs, wantArgs)
			}
		})
	}
}

func TestCheckImageWith_UnsupportedRuntime(t *testing.T) {
	t.Parallel()

	called := false
	output := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		called = true

		return nil, nil
	}

	err := checkImageWith(t.Context(), output, "nerdctl")
	if !errors.Is(err, ErrUnsupportedRuntime) {
		t.Errorf("checkImageWith = %v, want %v", err, ErrUnsupportedRuntime)
	}

	if called {
		t.Error("output was called for an unsupported runtime")
	}
}
