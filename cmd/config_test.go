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
	"testing"

	"github.com/Warashi/chelly/cmd"
)

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

	if cfg.ContainerSetupCmd != "" {
		t.Errorf("ContainerSetupCmd: got %q, want empty", cfg.ContainerSetupCmd)
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
container_setup_cmd = "echo setup"
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

	if len(cfg.AdditionalMounts) != 1 || cfg.AdditionalMounts[0] != "/host:/container" {
		t.Errorf("AdditionalMounts: got %v, want [/host:/container]", cfg.AdditionalMounts)
	}

	if cfg.ContainerSetupCmd != testSetupCmd {
		t.Errorf("ContainerSetupCmd: got %q, want %q", cfg.ContainerSetupCmd, testSetupCmd)
	}
}

func TestLoadConfigFrom_EnvVarOverridesConfigFile(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, dir, `
container_cmd = "docker"
config_home = "/config-file-context"
workdir = "/config-file-workdir"
additional_mounts = ["/config-file:/config-file"]
container_setup_cmd = "echo config-file"
`)

	t.Setenv("CHELLY_CONTAINER_CMD", testContainerCmdPodman)
	t.Setenv("CHELLY_CONFIG_HOME", "/env-context")
	t.Setenv("CHELLY_WORKDIR", "/env-workdir")
	t.Setenv("CHELLY_ADDITIONAL_MOUNTS", "/env-host:/env-container")
	t.Setenv("CHELLY_CONTAINER_SETUP_CMD", "echo env")

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

	if len(cfg.AdditionalMounts) != 1 || cfg.AdditionalMounts[0] != "/env-host:/env-container" {
		t.Errorf("AdditionalMounts: got %v, want [/env-host:/env-container]", cfg.AdditionalMounts)
	}

	if cfg.ContainerSetupCmd != "echo env" {
		t.Errorf("ContainerSetupCmd: got %q, want %q", cfg.ContainerSetupCmd, "echo env")
	}
}
