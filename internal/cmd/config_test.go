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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Warashi/chelly/internal/cmd"
	"github.com/Warashi/chelly/internal/config"
)

const (
	testContainerCmdPodman = "podman"
	testWorkspace          = "/workspace"
	testExtraArg           = "--userns=keep-id"
	cmdConfig              = "config"
	cmdConfigGet           = "get"
	cmdConfigSet           = "set"
)

func TestConfigGetCommand(t *testing.T) {
	configHome := t.TempDir()
	chellyDir := filepath.Join(configHome, "chelly")
	writeConfigFile(t, chellyDir, `container_cmd = "podman"`)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var out bytes.Buffer

	rootCmd := cmd.NewRootCommand()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{cmdConfig, cmdConfigGet, "container_cmd"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := out.String(); got != testContainerCmdPodman+"\n" {
		t.Errorf("output: got %q, want %q", got, testContainerCmdPodman+"\n")
	}
}

func TestConfigSetCommandUsesDefaultConfigDir(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var out bytes.Buffer

	rootCmd := cmd.NewRootCommand()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{cmdConfig, cmdConfigSet, "workdir", testWorkspace})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	chellyDir := filepath.Join(configHome, "chelly")

	cfg, err := config.LoadConfigFrom(chellyDir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.Workdir != testWorkspace {
		t.Errorf("Workdir: got %q, want %q", cfg.Workdir, testWorkspace)
	}
}

func TestConfigSetGetCommandEnvFiles(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var out bytes.Buffer

	rootCmd := cmd.NewRootCommand()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{cmdConfig, cmdConfigSet, "env_files", ".env,/etc/chelly/extra.env"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	rootCmd = cmd.NewRootCommand()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{cmdConfig, cmdConfigGet, "env_files"})

	out.Reset()

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got, want := out.String(), ".env,/etc/chelly/extra.env\n"; got != want {
		t.Errorf("output: got %q, want %q", got, want)
	}
}

func TestConfigSetCommandRuntimeOptionsBuild(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var out bytes.Buffer

	rootCmd := cmd.NewRootCommand()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{cmdConfig, cmdConfigSet, "--", "runtime_options.podman.build", testExtraArg})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	chellyDir := filepath.Join(configHome, "chelly")

	rootCmd = cmd.NewRootCommand()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{cmdConfig, cmdConfigGet, "runtime_options.podman.build"})

	out.Reset()

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := out.String(); got != testExtraArg+"\n" {
		t.Errorf("output: got %q, want %q", got, testExtraArg+"\n")
	}

	cfg, err := config.LoadConfigFrom(chellyDir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	want := []config.RuntimeOption{
		{Runtime: "podman", Subcommand: config.SubcommandBuild, Args: []string{testExtraArg}},
	}
	if !reflect.DeepEqual(cfg.RuntimeOptions, want) {
		t.Errorf("RuntimeOptions: got %v, want %v", cfg.RuntimeOptions, want)
	}
}

func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
