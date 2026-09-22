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

package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Warashi/chelly/internal/cmd"
	"github.com/Warashi/chelly/internal/container"
)

const cmdRun = "run"

// fakePodman installs a script named podman that logs its arguments and runs body.
func fakePodman(t *testing.T, body string) (string, string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, testContainerCmdPodman)
	logFile := filepath.Join(dir, "calls.log")
	script := "#!/bin/sh\necho \"$*\" >> " + logFile + "\n" + body + "\n"

	if err := os.WriteFile(path, []byte(script), 0o700); err != nil { //nolint:gosec
		t.Fatalf("WriteFile: %v", err)
	}

	return path, logFile
}

// runChelly executes `chelly run args...` and returns its error; stdout must stay
// empty because ACP-style callers own it.
func runChelly(t *testing.T, args ...string) error {
	t.Helper()

	var stdout bytes.Buffer

	rootCmd := cmd.NewRootCommand()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs(append([]string{cmdRun}, args...))

	err := rootCmd.Execute()

	if stdout.Len() != 0 {
		t.Errorf("stdout must stay empty, got %q", stdout.String())
	}

	if err != nil {
		return fmt.Errorf("executing run: %w", err)
	}

	return nil
}

func readCalls(t *testing.T, logFile string) string {
	t.Helper()

	data, err := os.ReadFile(logFile) //nolint:gosec // test-owned temp path
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadFile: %v", err)
	}

	return string(data)
}

func TestRunCommand_MissingImageFailsBeforeExec(t *testing.T) {
	podman, logFile := fakePodman(t, "exit 0")
	configHome := t.TempDir()
	writeConfigFile(t, filepath.Join(configHome, "chelly"), `container_cmd = "`+podman+`"`)

	t.Setenv("XDG_CONFIG_HOME", configHome)

	err := runChelly(t, "--", "true")
	if !errors.Is(err, container.ErrImageNotFound) {
		t.Fatalf("Execute = %v, want %v", err, container.ErrImageNotFound)
	}

	if !strings.Contains(err.Error(), "chelly build") {
		t.Errorf("error %q does not mention chelly build", err)
	}

	calls := readCalls(t, logFile)
	if !strings.Contains(calls, "image ls -q "+container.ImageName) {
		t.Errorf("calls %q lack the image check", calls)
	}

	if strings.Contains(calls, cmdRun+" ") {
		t.Errorf("calls %q reached run despite the missing image", calls)
	}
}

func TestRunCommand_CheckFailureIsNotReportedAsMissing(t *testing.T) {
	podman, _ := fakePodman(t, "echo 'cannot connect' >&2; exit 125")
	configHome := t.TempDir()
	writeConfigFile(t, filepath.Join(configHome, "chelly"), `container_cmd = "`+podman+`"`)

	t.Setenv("XDG_CONFIG_HOME", configHome)

	err := runChelly(t)
	if err == nil {
		t.Fatal("Execute returned nil")
	}

	if errors.Is(err, container.ErrImageNotFound) {
		t.Errorf("Execute = %v, must not be reported as a missing image", err)
	}
}

func TestRunCommand_RejectsPullInRuntimeOptions(t *testing.T) {
	podman, logFile := fakePodman(t, "echo abc123")
	configHome := t.TempDir()
	writeConfigFile(t, filepath.Join(configHome, "chelly"), `
container_cmd = "`+podman+`"

[[runtime_options]]
runtime = "podman"
subcommand = "run"
args = ["--pull=always"]
`)

	t.Setenv("XDG_CONFIG_HOME", configHome)

	err := runChelly(t)
	if !errors.Is(err, container.ErrPullOptionConflict) {
		t.Fatalf("Execute = %v, want %v", err, container.ErrPullOptionConflict)
	}

	if calls := readCalls(t, logFile); calls != "" {
		t.Errorf("runtime was called: %q", calls)
	}
}
