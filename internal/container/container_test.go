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

package container_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/Warashi/chelly/internal/container"
)

type (
	BuildConfig   = container.BuildConfig
	RunConfig     = container.RunConfig
	PodmanOptions = container.PodmanOptions
)

const (
	commandBuild           = "build"
	testContainerCmdDocker = "docker"
	testContainerCmdPodman = "podman"
	testConfigHome         = "/config/chelly"
	testWorkDir            = "/current/dir"
	testWorkDirMount       = "/current/dir:/current/dir"
	testWorkspace          = "/workspace"
	testSetupCmd           = "echo setup"
	testSetupCmd2          = "echo setup2"
	testInheritEnv         = "SSH_AUTH_SOCK"
	testInheritEnv2        = "GITHUB_TOKEN"
	testPodmanRunOption    = "--userns=keep-id"
	testPodmanRunOption2   = "--security-opt=label=disable"

	flagRM          = "--rm"
	flagVolume      = "--volume"
	flagEnv         = "--env"
	flagWorkdir     = "--workdir"
	flagInteractive = "--interactive"
	flagTTY         = "--tty"
	shellSh         = "sh"
	shellFlagLC     = "-lc"
	cmdRun          = "run"
	cmdEcho         = "echo"
	cmdBash         = "bash"
	cmdHello        = "hello"
)

func baseBuildConfig() BuildConfig {
	return BuildConfig{
		ConfigHome: testConfigHome,
	}
}

func baseRunConfig() RunConfig {
	return RunConfig{
		ContainerCmd:       testContainerCmdDocker,
		Workdir:            testWorkDir,
		AdditionalMounts:   nil,
		ContainerSetupCmds: nil,
		InheritEnv:         nil,
		PodmanOptions:      PodmanOptions{Run: nil},
	}
}

func TestBuildArgs_WithoutNoCache(t *testing.T) {
	t.Parallel()

	cfg := baseBuildConfig()
	got := container.BuildArgs(cfg, false)
	want := []string{commandBuild, "--tag", container.ImageName, testConfigHome}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildArgs: got %v, want %v", got, want)
	}
}

func TestBuildArgs_WithNoCache(t *testing.T) {
	t.Parallel()

	cfg := baseBuildConfig()
	got := container.BuildArgs(cfg, true)
	want := []string{commandBuild, "--no-cache", "--tag", container.ImageName, testConfigHome}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_Default(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	got := container.RunArgs(cfg, testWorkDir, false, []string{cmdEcho, cmdHello})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		cmdEcho, cmdHello,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_WithTTY(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	got := container.RunArgs(cfg, testWorkDir, true, []string{cmdBash})
	want := []string{
		cmdRun, flagRM,
		flagInteractive, flagTTY,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		cmdBash,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_AdditionalMounts(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.AdditionalMounts = []string{"/host1:/cont1", "/host2:/cont2"}

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagVolume, "/host1:/cont1",
		flagVolume, "/host2:/cont2",
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_PodmanOptionsRun(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerCmd = testContainerCmdPodman
	cfg.PodmanOptions.Run = []string{testPodmanRunOption, testPodmanRunOption2}

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		testPodmanRunOption, testPodmanRunOption2,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_InheritEnv(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.InheritEnv = []string{testInheritEnv, testInheritEnv2}

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagEnv, testInheritEnv,
		flagEnv, testInheritEnv2,
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_InheritEnvAfterPodmanOptions(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerCmd = testContainerCmdPodman
	cfg.PodmanOptions.Run = []string{"--env", "SSH_AUTH_SOCK=/tmp/socket"}
	cfg.InheritEnv = []string{testInheritEnv}

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		"--env", "SSH_AUTH_SOCK=/tmp/socket",
		flagVolume, testWorkDirMount,
		flagEnv, testInheritEnv,
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_PodmanOptionsRunWithPath(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerCmd = "/usr/bin/podman"
	cfg.PodmanOptions.Run = []string{testPodmanRunOption}

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		testPodmanRunOption,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_PodmanOptionsRunIgnoredForDocker(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.PodmanOptions.Run = []string{testPodmanRunOption}

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_SetupCmdWithCommand(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerSetupCmds = []string{testSetupCmd}

	got := container.RunArgs(cfg, testWorkDir, false, []string{cmdEcho, cmdHello})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		shellSh, shellFlagLC, `echo setup >&2 && exec "$@"`, shellSh,
		cmdEcho, cmdHello,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_SetupCmdWithoutCommand(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerSetupCmds = []string{testSetupCmd}

	got := container.RunArgs(cfg, testWorkDir, false, []string{})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		shellSh, shellFlagLC, "echo setup >&2",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_MultipleSetupCmdsWithCommand(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerSetupCmds = []string{testSetupCmd, testSetupCmd2}

	got := container.RunArgs(cfg, testWorkDir, false, []string{cmdEcho, cmdHello})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		shellSh, shellFlagLC, `echo setup >&2 & p0=$!; echo setup2 >&2 & p1=$!; wait $p0 && wait $p1 && exec "$@"`, shellSh,
		cmdEcho, cmdHello,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_MultipleSetupCmdsWithoutCommand(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerSetupCmds = []string{testSetupCmd, testSetupCmd2}

	got := container.RunArgs(cfg, testWorkDir, false, []string{})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		shellSh, shellFlagLC, "echo setup >&2 & p0=$!; echo setup2 >&2 & p1=$!; wait $p0 && wait $p1",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_CustomWorkdir(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.Workdir = testWorkspace

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkspace,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_AllOptions(t *testing.T) {
	t.Parallel()

	cfg := RunConfig{
		ContainerCmd:       testContainerCmdPodman,
		Workdir:            testWorkspace,
		AdditionalMounts:   []string{"/home/user/.ssh:/home/user/.ssh"},
		ContainerSetupCmds: []string{"source /etc/profile"},
		InheritEnv:         []string{testInheritEnv},
		PodmanOptions:      PodmanOptions{Run: []string{testPodmanRunOption}},
	}

	got := container.RunArgs(cfg, testWorkDir, false, []string{cmdBash, "-c", "echo hi"})
	want := []string{
		cmdRun, flagRM,
		testPodmanRunOption,
		flagVolume, testWorkDirMount,
		flagVolume, "/home/user/.ssh:/home/user/.ssh",
		flagEnv, testInheritEnv,
		flagWorkdir, testWorkspace,
		container.ImageName,
		shellSh, shellFlagLC, `source /etc/profile >&2 && exec "$@"`, shellSh,
		cmdBash, "-c", "echo hi",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs (all options): got %v, want %v", got, want)
	}
}

func TestStripDashDash(t *testing.T) {
	t.Parallel()

	got := container.StripDashDash([]string{"--", "echo", "hello"})
	want := []string{"echo", "hello"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("StripDashDash: got %v, want %v", got, want)
	}
}

func TestIsTTY_RegularFileFalse(t *testing.T) {
	t.Parallel()

	file, err := os.Open("../../go.mod")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	if container.IsTTY(file) {
		t.Error("IsTTY returned true for regular file")
	}
}

func TestRun_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	err := container.Run(context.Background(), "false", nil)
	if err == nil {
		t.Fatal("Run returned nil")
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Run error = %v, want wrapped *exec.ExitError", err)
	}
}
