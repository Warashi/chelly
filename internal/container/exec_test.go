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

//nolint:testpackage // Tests execWith seams to avoid replacing the test process.
package container

import (
	"errors"
	"slices"
	"testing"
)

var (
	errExecveFailed = errors.New("execve failed")
	errLookPath     = errors.New("not found")
)

func TestExecWith_UsesLookPathExecveAndEnvironment(t *testing.T) {
	t.Parallel()

	expectedPath := "/usr/bin/docker"
	expectedArgv := []string{"docker", "run", "chelly:latest"}
	expectedEnv := []string{"CHELLY_TEST=1"}

	var (
		gotLookPath string
		gotPath     string
		gotArgv     []string
		gotEnv      []string
	)

	deps := execDeps{
		lookPath: func(command string) (string, error) {
			gotLookPath = command

			return expectedPath, nil
		},
		execve: func(path string, argv []string, env []string) error {
			gotPath = path

			gotArgv = append([]string(nil), argv...)
			gotEnv = append([]string(nil), env...)

			return errExecveFailed
		},
		environ: func() []string {
			return expectedEnv
		},
	}

	err := execWith(deps, "docker", expectedArgv[1:])
	if !errors.Is(err, errExecveFailed) {
		t.Fatalf("execWith error = %v, want wrapped execveErr", err)
	}

	if gotLookPath != expectedArgv[0] {
		t.Errorf("lookPath command = %q, want %q", gotLookPath, expectedArgv[0])
	}

	if gotPath != expectedPath {
		t.Errorf("execve path = %q, want %q", gotPath, expectedPath)
	}

	if !slices.Equal(gotArgv, expectedArgv) {
		t.Errorf("execve argv = %v, want %v", gotArgv, expectedArgv)
	}

	if !slices.Equal(gotEnv, expectedEnv) {
		t.Errorf("execve env = %v, want %v", gotEnv, expectedEnv)
	}
}

func TestExecWith_ReturnsLookPathError(t *testing.T) {
	t.Parallel()

	deps := execDeps{
		lookPath: func(string) (string, error) {
			return "", errLookPath
		},
		execve: func(string, []string, []string) error {
			t.Fatal("execve called after lookPath failed")

			return nil
		},
		environ: func() []string {
			t.Fatal("environ called after lookPath failed")

			return nil
		},
	}

	err := execWith(deps, "missing-container-command", nil)
	if !errors.Is(err, errLookPath) {
		t.Fatalf("execWith error = %v, want wrapped lookPathErr", err)
	}
}
