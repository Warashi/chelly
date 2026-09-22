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

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

const (
	commandBuild = "build"

	// ImageName is the chelly container image tag.
	ImageName = "chelly:latest"
)

// BuildConfig holds configuration used by the container image build.
type BuildConfig struct {
	ConfigHome string
	ExtraArgs  []string
}

// RunConfig holds configuration used by container execution.
type RunConfig struct {
	ContainerCmd     string
	Workdir          string
	AdditionalMounts []string
	InheritEnv       []string
	EnvFiles         []string
	ExtraArgs        []string
}

// BuildArgs returns the argument slice for the container build command.
func BuildArgs(cfg BuildConfig, noCacheFlag bool) []string {
	args := []string{commandBuild}
	if noCacheFlag {
		args = append(args, "--no-cache")
	}

	args = append(args, cfg.ExtraArgs...)
	args = append(args, "--tag", ImageName, cfg.ConfigHome)

	return args
}

// IsTTY reports whether f is connected to a terminal.
func IsTTY(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// RunArgs returns the argument slice for the container run command.
func RunArgs(cfg RunConfig, workDir string, isTTY bool, userArgs []string, autoMounts ...string) []string {
	args := []string{"run", "--rm", "--interactive"}

	if isTTY {
		args = append(args, "--tty")
	}

	// The image entrypoint exec-chains into the user command, which then sits at
	// PID 1 and only waits for its own children. Orphans it inherits would stay
	// zombies for the whole session, so every run asks the runtime for an init
	// (podman, docker, and Apple container all spell it `--init`).
	args = append(args, "--init")
	args = append(args, pullPolicyArgs(cfg.ContainerCmd)...)
	args = append(args, cfg.ExtraArgs...)

	seenMounts := map[string]struct{}{}
	args = appendMount(args, seenMounts, workDir+":"+workDir)

	for _, mount := range autoMounts {
		args = appendMount(args, seenMounts, mount+":"+mount)
	}

	for _, mount := range cfg.AdditionalMounts {
		args = appendMount(args, seenMounts, mount)
	}

	for _, file := range cfg.EnvFiles {
		args = append(args, "--env-file", file)
	}

	for _, name := range cfg.InheritEnv {
		args = append(args, "--env", name)
	}

	args = append(args, "--workdir", cfg.Workdir)
	args = append(args, ImageName)
	args = append(args, userArgs...)

	return args
}

func appendMount(args []string, seen map[string]struct{}, mount string) []string {
	if _, ok := seen[mount]; ok {
		return args
	}

	seen[mount] = struct{}{}

	return append(args, "--volume", mount)
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
