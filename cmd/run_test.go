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
	"reflect"
	"testing"

	"github.com/Warashi/chelly/cmd"
)

func runConfig() cmd.Config {
	return cmd.Config{
		ContainerCmd:      testContainerCmdDocker,
		ConfigHome:        testConfigHome,
		Workdir:           testWorkDir,
		AdditionalMounts:  nil,
		ContainerSetupCmd: "",
	}
}

func TestContainerRunArgs_Default(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, false, []string{cmdEcho, cmdHello})
	want := []string{
		cmdRun, flagRM,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		imageChelly,
		cmdEcho, cmdHello,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_WithTTY(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, true, []string{cmdBash})
	want := []string{
		cmdRun, flagRM,
		flagInteractive, flagTTY,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		imageChelly,
		cmdBash,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_ForceTTY(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	got := cmd.ContainerRunArgs(cfg, testWorkDir, true, false, []string{shellSh})
	want := []string{
		cmdRun, flagRM,
		flagInteractive, flagTTY,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		imageChelly,
		shellSh,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_AdditionalMounts(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	cfg.AdditionalMounts = []string{"/host1:/cont1", "/host2:/cont2"}

	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagVolume, "/host1:/cont1",
		flagVolume, "/host2:/cont2",
		flagWorkdir, testWorkDir,
		imageChelly,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_SetupCmdWithCommand(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	cfg.ContainerSetupCmd = testSetupCmd

	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, false, []string{cmdEcho, cmdHello})
	want := []string{
		cmdRun, flagRM,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		imageChelly,
		shellSh, shellFlagLC, `echo setup; exec "$@"`, shellSh,
		cmdEcho, cmdHello,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_SetupCmdWithoutCommand(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	cfg.ContainerSetupCmd = testSetupCmd

	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, false, []string{})
	want := []string{
		cmdRun, flagRM,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		imageChelly,
		shellSh, shellFlagLC, testSetupCmd,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_CustomWorkdir(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	cfg.Workdir = testWorkspace

	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkspace,
		imageChelly,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs: got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_RunitDefaultsToSh(t *testing.T) {
	t.Parallel()

	cfg := runConfig()
	userArgs := []string{shellSh}
	got := cmd.ContainerRunArgs(cfg, testWorkDir, true, false, userArgs)
	want := []string{
		cmdRun, flagRM,
		flagInteractive, flagTTY,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		imageChelly,
		shellSh,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs (runit): got %v, want %v", got, want)
	}
}

func TestContainerRunArgs_AllOptions(t *testing.T) {
	t.Parallel()

	cfg := cmd.Config{
		ContainerCmd:      testContainerCmdPodman,
		ConfigHome:        testConfigHome,
		Workdir:           testWorkspace,
		AdditionalMounts:  []string{"/home/user/.ssh:/home/user/.ssh"},
		ContainerSetupCmd: "source /etc/profile",
	}

	got := cmd.ContainerRunArgs(cfg, testWorkDir, false, false, []string{cmdBash, "-c", "echo hi"})
	want := []string{
		cmdRun, flagRM,
		flagVolume, volumeHomeMount,
		flagVolume, testWorkDirMount,
		flagVolume, "/home/user/.ssh:/home/user/.ssh",
		flagWorkdir, testWorkspace,
		imageChelly,
		shellSh, shellFlagLC, `source /etc/profile; exec "$@"`, shellSh,
		cmdBash, "-c", "echo hi",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ContainerRunArgs (all options): got %v, want %v", got, want)
	}
}
