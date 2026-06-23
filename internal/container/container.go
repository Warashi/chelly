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

// Package container builds command arguments and executes the chelly container.
package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	commandBuild = "build"

	// ImageName is the chelly container image tag.
	ImageName = "chelly:latest"
)

// PodmanOptions holds Podman-specific container options.
type PodmanOptions struct {
	Run []string
}

// BuildConfig holds configuration used by the container image build.
type BuildConfig struct {
	ConfigHome string
}

// RunConfig holds configuration used by container execution.
type RunConfig struct {
	ContainerCmd       string
	Workdir            string
	AdditionalMounts   []string
	ContainerSetupCmds []string
	InheritEnv         []string
	PodmanOptions      PodmanOptions
}

// BuildArgs returns the argument slice for the container build command.
func BuildArgs(cfg BuildConfig, noCacheFlag bool) []string {
	args := []string{commandBuild}
	if noCacheFlag {
		args = append(args, "--no-cache")
	}

	args = append(args, "--tag", ImageName, cfg.ConfigHome)

	return args
}

// IsTTY reports whether f is connected to a terminal.
func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

// RunArgs returns the argument slice for the container run command.
func RunArgs(cfg RunConfig, workDir string, isTTY bool, userArgs []string) []string {
	args := []string{"run", "--rm"}

	if isTTY {
		args = append(args, "--interactive", "--tty")
	}

	if filepath.Base(cfg.ContainerCmd) == "podman" {
		args = append(args, cfg.PodmanOptions.Run...)
	}

	args = append(args, "--volume", workDir+":"+workDir)

	for _, mount := range cfg.AdditionalMounts {
		args = append(args, "--volume", mount)
	}

	for _, name := range cfg.InheritEnv {
		args = append(args, "--env", name)
	}

	args = append(args, "--workdir", cfg.Workdir)
	args = append(args, ImageName)

	if len(cfg.ContainerSetupCmds) > 0 {
		script := buildSetupScript(cfg.ContainerSetupCmds, userArgs)
		args = append(args, "sh", "-lc", script)

		if len(userArgs) > 0 {
			args = append(args, "sh")
			args = append(args, userArgs...)
		}
	} else {
		args = append(args, userArgs...)
	}

	return args
}

func buildSetupScript(cmds []string, userArgs []string) string {
	execSuffix := ""
	if len(userArgs) > 0 {
		execSuffix = ` && exec "$@"`
	}

	if len(cmds) == 1 {
		return cmds[0] + " >&2" + execSuffix
	}

	var parts []string

	for i, c := range cmds {
		parts = append(parts, fmt.Sprintf("%s >&2 & p%d=$!", c, i))
	}

	var waits []string

	for i := range cmds {
		waits = append(waits, fmt.Sprintf("wait $p%d", i))
	}

	return strings.Join(parts, "; ") + "; " + strings.Join(waits, " && ") + execSuffix
}

// StripDashDash removes Cobra's command separator from user command arguments.
func StripDashDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}

	return args
}

type execDeps struct {
	lookPath func(string) (string, error)
	execve   func(string, []string, []string) error
	environ  func() []string
}

var defaultExecDeps = execDeps{
	lookPath: exec.LookPath,
	execve:   unix.Exec,
	environ:  os.Environ,
}

// Run runs the configured container command as a child process with connected standard streams.
func Run(ctx context.Context, containerCmd string, args []string) error {
	cmd := exec.CommandContext(ctx, containerCmd, args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running container: %w", err)
	}

	return nil
}

// Exec replaces the current process with the configured container command.
func Exec(containerCmd string, args []string) error {
	return execWith(defaultExecDeps, containerCmd, args)
}

func execWith(deps execDeps, containerCmd string, args []string) error {
	path, err := deps.lookPath(containerCmd)
	if err != nil {
		return fmt.Errorf("looking up container command: %w", err)
	}

	argv := append([]string{containerCmd}, args...)
	if err := deps.execve(path, argv, deps.environ()); err != nil {
		return fmt.Errorf("replacing process with container: %w", err)
	}

	return nil
}
