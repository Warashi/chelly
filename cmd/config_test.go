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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Warashi/chelly/cmd"
	"github.com/pelletier/go-toml/v2"
)

func assertStringSlice(t *testing.T, name string, got, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

func TestFormatConfig_RoundTrip(t *testing.T) {
	t.Parallel()

	original := cmd.Config{
		ContainerCmd:       testContainerCmdPodman,
		ConfigHome:         testConfigHome,
		Workdir:            testWorkspace,
		AdditionalMounts:   []string{testMountA},
		ContainerSetupCmds: []string{testSetupCmd},
		PodmanOptions:      cmd.PodmanOptions{Run: []string{testPodmanRunOption}},
	}

	out, err := cmd.FormatConfig(original)
	if err != nil {
		t.Fatalf("FormatConfig: %v", err)
	}

	var roundTripped cmd.Config
	if err := toml.Unmarshal([]byte(out), &roundTripped); err != nil {
		t.Fatalf("unmarshal round-trip: %v", err)
	}

	if roundTripped.ContainerCmd != original.ContainerCmd {
		t.Errorf("ContainerCmd: got %q, want %q", roundTripped.ContainerCmd, original.ContainerCmd)
	}

	if roundTripped.ConfigHome != original.ConfigHome {
		t.Errorf("ConfigHome: got %q, want %q", roundTripped.ConfigHome, original.ConfigHome)
	}

	if roundTripped.Workdir != original.Workdir {
		t.Errorf("Workdir: got %q, want %q", roundTripped.Workdir, original.Workdir)
	}

	assertStringSlice(t, "AdditionalMounts", roundTripped.AdditionalMounts, original.AdditionalMounts)
	assertStringSlice(t, "ContainerSetupCmds", roundTripped.ContainerSetupCmds, original.ContainerSetupCmds)
	assertStringSlice(t, "PodmanOptions.Run", roundTripped.PodmanOptions.Run, original.PodmanOptions.Run)
}

func TestGetConfigValue(t *testing.T) {
	t.Parallel()

	cfg := cmd.Config{
		ContainerCmd:       testContainerCmdPodman,
		ConfigHome:         testConfigHome,
		Workdir:            testWorkspace,
		AdditionalMounts:   []string{testMountA, testMountB},
		ContainerSetupCmds: []string{testSetupCmd, testSetupCmd2},
		PodmanOptions:      cmd.PodmanOptions{Run: []string{testPodmanRunOption, testPodmanRunOption2}},
	}

	cases := []struct {
		key  string
		want string
	}{
		{"container_cmd", testContainerCmdPodman},
		{"config_home", testConfigHome},
		{"workdir", testWorkspace},
		{"additional_mounts", testMountA + "," + testMountB},
		{"container_setup_cmds", testSetupCmd + "," + testSetupCmd2},
		{"podman_options.run", testPodmanRunOption + "," + testPodmanRunOption2},
	}

	for _, testCase := range cases {
		t.Run(testCase.key, func(t *testing.T) {
			t.Parallel()

			got, err := cmd.GetConfigValue(cfg, testCase.key)
			if err != nil {
				t.Fatalf("GetConfigValue(%q): %v", testCase.key, err)
			}

			if got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestGetConfigValue_UnknownKey(t *testing.T) {
	t.Parallel()

	_, err := cmd.GetConfigValue(cmd.Config{
		ContainerCmd:       "",
		ConfigHome:         "",
		Workdir:            "",
		AdditionalMounts:   nil,
		ContainerSetupCmds: nil,
		PodmanOptions:      cmd.PodmanOptions{Run: nil},
	}, "nonexistent_key")
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}

	if !strings.Contains(err.Error(), "nonexistent_key") {
		t.Errorf("error should mention the unknown key: %v", err)
	}
}

func TestSetConfigValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := cmd.SetConfigValue(dir, "container_cmd", testContainerCmdPodman); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.ContainerCmd != testContainerCmdPodman {
		t.Errorf("ContainerCmd: got %q, want %q", cfg.ContainerCmd, testContainerCmdPodman)
	}
}

func TestSetConfigValue_CreatesFileAndDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "newdir")

	if err := cmd.SetConfigValue(dir, "workdir", testWorkspace); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Errorf("config.toml not created: %v", err)
	}

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.Workdir != testWorkspace {
		t.Errorf("Workdir: got %q, want %q", cfg.Workdir, testWorkspace)
	}
}

func TestSetConfigValue_PreservesExistingKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeConfigFile(t, dir, `container_cmd = "docker"`)

	if err := cmd.SetConfigValue(dir, "workdir", testWorkspace); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.ContainerCmd != testContainerCmdDocker {
		t.Errorf("ContainerCmd: got %q, want %q", cfg.ContainerCmd, testContainerCmdDocker)
	}

	if cfg.Workdir != testWorkspace {
		t.Errorf("Workdir: got %q, want %q", cfg.Workdir, testWorkspace)
	}
}

func TestSetConfigValue_AdditionalMounts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  []string
	}{
		{"single", testMountA, []string{testMountA}},
		{"multiple", testMountA + "," + testMountB, []string{testMountA, testMountB}},
		{"with spaces", " " + testMountA + " , " + testMountB + " ", []string{testMountA, testMountB}},
		{"empty", "", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			if err := cmd.SetConfigValue(dir, "additional_mounts", testCase.value); err != nil {
				t.Fatalf("SetConfigValue: %v", err)
			}

			cfg, err := cmd.LoadConfigFrom(dir)
			if err != nil {
				t.Fatalf("LoadConfigFrom: %v", err)
			}

			if len(cfg.AdditionalMounts) != len(testCase.want) {
				t.Errorf("AdditionalMounts len: got %d, want %d", len(cfg.AdditionalMounts), len(testCase.want))

				return
			}

			for i, m := range testCase.want {
				if cfg.AdditionalMounts[i] != m {
					t.Errorf("AdditionalMounts[%d]: got %q, want %q", i, cfg.AdditionalMounts[i], m)
				}
			}
		})
	}
}

func TestSetConfigValue_PodmanOptionsRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := cmd.SetConfigValue(dir, "podman_options.run", testPodmanRunOption+","+testPodmanRunOption2); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if len(cfg.PodmanOptions.Run) != 2 ||
		cfg.PodmanOptions.Run[0] != testPodmanRunOption ||
		cfg.PodmanOptions.Run[1] != testPodmanRunOption2 {
		t.Errorf("PodmanOptions.Run: got %v, want [%q %q]", cfg.PodmanOptions.Run, testPodmanRunOption, testPodmanRunOption2)
	}
}

func TestSetConfigValue_UnknownKey(t *testing.T) {
	t.Parallel()

	err := cmd.SetConfigValue(t.TempDir(), "nonexistent_key", "value")
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}

	if !strings.Contains(err.Error(), "nonexistent_key") {
		t.Errorf("error should mention the unknown key: %v", err)
	}
}

func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestLoadConfigFrom_Defaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.ConfigHome != dir {
		t.Errorf("ConfigHome: got %q, want %q", cfg.ConfigHome, dir)
	}

	if len(cfg.AdditionalMounts) != 0 {
		t.Errorf("AdditionalMounts: got %v, want empty", cfg.AdditionalMounts)
	}

	if len(cfg.ContainerSetupCmds) != 0 {
		t.Errorf("ContainerSetupCmds: got %v, want empty", cfg.ContainerSetupCmds)
	}

	if len(cfg.PodmanOptions.Run) != 0 {
		t.Errorf("PodmanOptions.Run: got %v, want empty", cfg.PodmanOptions.Run)
	}

	if cfg.Workdir == "" {
		t.Error("Workdir: got empty, want current directory")
	}
}

func TestLoadConfigFrom_ConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeConfigFile(t, dir, `
container_cmd = "podman"
config_home = "/custom/context"
workdir = "/workspace"
additional_mounts = ["/host:/container"]
container_setup_cmds = ["echo setup"]

[podman_options]
run = ["--userns=keep-id"]
`)

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.ContainerCmd != testContainerCmdPodman {
		t.Errorf("ContainerCmd: got %q, want %q", cfg.ContainerCmd, testContainerCmdPodman)
	}

	if cfg.ConfigHome != "/custom/context" {
		t.Errorf("ConfigHome: got %q, want %q", cfg.ConfigHome, "/custom/context")
	}

	if cfg.Workdir != testWorkspace {
		t.Errorf("Workdir: got %q, want %q", cfg.Workdir, testWorkspace)
	}

	assertStringSlice(t, "AdditionalMounts", cfg.AdditionalMounts, []string{"/host:/container"})
	assertStringSlice(t, "ContainerSetupCmds", cfg.ContainerSetupCmds, []string{testSetupCmd})
	assertStringSlice(t, "PodmanOptions.Run", cfg.PodmanOptions.Run, []string{testPodmanRunOption})
}

func TestLoadConfig_XDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	chellyDir := filepath.Join(dir, "chelly")
	writeConfigFile(t, chellyDir, `
container_cmd = "podman"
`)

	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := cmd.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.ContainerCmd != testContainerCmdPodman {
		t.Errorf("ContainerCmd: got %q, want %q", cfg.ContainerCmd, testContainerCmdPodman)
	}

	if cfg.ConfigHome != chellyDir {
		t.Errorf("ConfigHome: got %q, want %q", cfg.ConfigHome, chellyDir)
	}
}

func TestLoadConfigFrom_EnvVarOverridesConfigFile(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, dir, `
container_cmd = "docker"
config_home = "/config-file-context"
workdir = "/config-file-workdir"
additional_mounts = ["/config-file:/config-file"]
container_setup_cmds = ["echo config-file"]

[podman_options]
run = ["--userns=keep-id"]
`)

	t.Setenv("CHELLY_CONTAINER_CMD", testContainerCmdPodman)
	t.Setenv("CHELLY_CONFIG_HOME", "/env-context")
	t.Setenv("CHELLY_WORKDIR", "/env-workdir")
	t.Setenv("CHELLY_ADDITIONAL_MOUNTS", "/env-host:/env-container")
	t.Setenv("CHELLY_CONTAINER_SETUP_CMDS", "echo-env")
	t.Setenv("CHELLY_PODMAN_OPTIONS_RUN", testPodmanRunOption+","+testPodmanRunOption2)

	cfg, err := cmd.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.ContainerCmd != testContainerCmdPodman {
		t.Errorf("ContainerCmd: got %q, want %q", cfg.ContainerCmd, testContainerCmdPodman)
	}

	if cfg.ConfigHome != "/env-context" {
		t.Errorf("ConfigHome: got %q, want %q", cfg.ConfigHome, "/env-context")
	}

	if cfg.Workdir != "/env-workdir" {
		t.Errorf("Workdir: got %q, want %q", cfg.Workdir, "/env-workdir")
	}

	assertStringSlice(t, "AdditionalMounts", cfg.AdditionalMounts, []string{"/env-host:/env-container"})
	assertStringSlice(t, "ContainerSetupCmds", cfg.ContainerSetupCmds, []string{"echo-env"})
	assertStringSlice(t, "PodmanOptions.Run", cfg.PodmanOptions.Run, []string{testPodmanRunOption, testPodmanRunOption2})
}
