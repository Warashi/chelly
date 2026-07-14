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

package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Warashi/chelly/internal/config"
	"github.com/pelletier/go-toml/v2"
)

const (
	testContainerCmdDocker = "docker"
	testContainerCmdPodman = "podman"
	testConfigHome         = "/config/chelly"
	testWorkspace          = "/workspace"
	testMountA             = "/a:/a"
	testMountB             = "/b:/b"
	testInheritEnv         = "SSH_AUTH_SOCK"
	testInheritEnv2        = "GITHUB_TOKEN"
	testEnvFile            = ".env"
	testEnvFile2           = "/etc/chelly/extra.env"
	testExtraArg           = "--userns=keep-id"
	testExtraArg2          = "--security-opt=label=disable"
	testRuntimePodman      = "podman"
	testRuntimeDocker      = "docker"
	keyRuntimeOptionsRun   = "runtime_options.podman.run"
	keyRuntimeOptionsBuild = "runtime_options.podman.build"
	testCaseNameEmpty      = "empty"
)

func assertStringSlice(t *testing.T, name string, got, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

func TestFormatConfig_RoundTrip(t *testing.T) {
	t.Parallel()

	original := config.Config{
		ContainerCmd:     testContainerCmdPodman,
		ConfigHome:       testConfigHome,
		Workdir:          testWorkspace,
		AdditionalMounts: []string{testMountA},
		InheritEnv:       []string{testInheritEnv},
		EnvFiles:         []string{testEnvFile},
		RuntimeOptions: []config.RuntimeOption{
			{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
		},
	}

	out, err := config.FormatConfig(original)
	if err != nil {
		t.Fatalf("FormatConfig: %v", err)
	}

	var roundTripped config.Config
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
	assertStringSlice(t, "InheritEnv", roundTripped.InheritEnv, original.InheritEnv)
	assertStringSlice(t, "EnvFiles", roundTripped.EnvFiles, original.EnvFiles)

	if !reflect.DeepEqual(roundTripped.RuntimeOptions, original.RuntimeOptions) {
		t.Errorf("RuntimeOptions: got %v, want %v", roundTripped.RuntimeOptions, original.RuntimeOptions)
	}
}

func TestGetConfigValue(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		ContainerCmd:     testContainerCmdPodman,
		ConfigHome:       testConfigHome,
		Workdir:          testWorkspace,
		AdditionalMounts: []string{testMountA, testMountB},
		InheritEnv:       []string{testInheritEnv, testInheritEnv2},
		EnvFiles:         []string{testEnvFile, testEnvFile2},
		RuntimeOptions: []config.RuntimeOption{
			{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg, testExtraArg2}},
		},
	}

	cases := []struct {
		key  string
		want string
	}{
		{"container_cmd", testContainerCmdPodman},
		{"config_home", testConfigHome},
		{"workdir", testWorkspace},
		{"additional_mounts", testMountA + "," + testMountB},
		{"inherit_env", testInheritEnv + "," + testInheritEnv2},
		{"env_files", testEnvFile + "," + testEnvFile2},
		{keyRuntimeOptionsRun, testExtraArg + "," + testExtraArg2},
	}

	for _, testCase := range cases {
		t.Run(testCase.key, func(t *testing.T) {
			t.Parallel()

			got, err := config.GetConfigValue(cfg, testCase.key)
			if err != nil {
				t.Fatalf("GetConfigValue(%q): %v", testCase.key, err)
			}

			if got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestGetConfigValue_RuntimeOptionsUnsetReturnsEmpty(t *testing.T) {
	t.Parallel()

	got, err := config.GetConfigValue(config.Config{
		ContainerCmd:     "",
		ConfigHome:       "",
		Workdir:          "",
		AdditionalMounts: nil,
		InheritEnv:       nil,
		EnvFiles:         nil,
		RuntimeOptions:   nil,
	}, keyRuntimeOptionsBuild)
	if err != nil {
		t.Fatalf("GetConfigValue: %v", err)
	}

	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestGetConfigValue_RuntimeOptionsInvalidSubcommand(t *testing.T) {
	t.Parallel()

	_, err := config.GetConfigValue(config.Config{
		ContainerCmd:     "",
		ConfigHome:       "",
		Workdir:          "",
		AdditionalMounts: nil,
		InheritEnv:       nil,
		EnvFiles:         nil,
		RuntimeOptions:   nil,
	}, "runtime_options.podman.exec")
	if err == nil {
		t.Fatal("expected error for invalid subcommand, got nil")
	}

	if !strings.Contains(err.Error(), "exec") {
		t.Errorf("error should mention the invalid subcommand: %v", err)
	}
}

func TestGetConfigValue_UnknownKey(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"nonexistent_key", "container_setup_cmds"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			_, err := config.GetConfigValue(config.Config{
				ContainerCmd:     "",
				ConfigHome:       "",
				Workdir:          "",
				AdditionalMounts: nil,
				InheritEnv:       nil,
				EnvFiles:         nil,
				RuntimeOptions:   nil,
			}, key)
			if err == nil {
				t.Fatal("expected error for unknown key, got nil")
			}

			if !strings.Contains(err.Error(), key) {
				t.Errorf("error should mention the unknown key: %v", err)
			}
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := config.SetConfigValue(dir, "container_cmd", testContainerCmdPodman); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
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

	if err := config.SetConfigValue(dir, "workdir", testWorkspace); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Errorf("config.toml not created: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
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

	if err := config.SetConfigValue(dir, "workdir", testWorkspace); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
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
		{testCaseNameEmpty, "", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			if err := config.SetConfigValue(dir, "additional_mounts", testCase.value); err != nil {
				t.Fatalf("SetConfigValue: %v", err)
			}

			cfg, err := config.LoadConfigFrom(dir)
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

func TestSetConfigValue_RuntimeOptions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := config.SetConfigValue(dir, keyRuntimeOptionsRun, testExtraArg+","+testExtraArg2); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	want := []config.RuntimeOption{
		{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg, testExtraArg2}},
	}

	if !reflect.DeepEqual(cfg.RuntimeOptions, want) {
		t.Errorf("RuntimeOptions: got %v, want %v", cfg.RuntimeOptions, want)
	}
}

func TestSetConfigValue_RuntimeOptionsUpdatesExistingEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := config.SetConfigValue(dir, keyRuntimeOptionsRun, testExtraArg); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	if err := config.SetConfigValue(dir, keyRuntimeOptionsBuild, testExtraArg2); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	if err := config.SetConfigValue(dir, keyRuntimeOptionsRun, testExtraArg2); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	want := []config.RuntimeOption{
		{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg2}},
		{Runtime: testRuntimePodman, Subcommand: config.SubcommandBuild, Args: []string{testExtraArg2}},
	}

	if !reflect.DeepEqual(cfg.RuntimeOptions, want) {
		t.Errorf("RuntimeOptions: got %v, want %v", cfg.RuntimeOptions, want)
	}
}

func TestSetConfigValue_RuntimeOptionsInvalidSubcommand(t *testing.T) {
	t.Parallel()

	err := config.SetConfigValue(t.TempDir(), "runtime_options.podman.exec", testExtraArg)
	if err == nil {
		t.Fatal("expected error for invalid subcommand, got nil")
	}

	if !strings.Contains(err.Error(), "exec") {
		t.Errorf("error should mention the invalid subcommand: %v", err)
	}
}

func TestSetConfigValue_InheritEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := config.SetConfigValue(dir, "inherit_env", testInheritEnv+","+testInheritEnv2); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	assertStringSlice(t, "InheritEnv", cfg.InheritEnv, []string{testInheritEnv, testInheritEnv2})
}

func TestSetConfigValue_EnvFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := config.SetConfigValue(dir, "env_files", testEnvFile+","+testEnvFile2); err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}

	cfg, err := config.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	assertStringSlice(t, "EnvFiles", cfg.EnvFiles, []string{testEnvFile, testEnvFile2})
}

func TestSetConfigValue_UnknownKey(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"nonexistent_key", "container_setup_cmds"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			err := config.SetConfigValue(t.TempDir(), key, "value")
			if err == nil {
				t.Fatal("expected error for unknown key, got nil")
			}

			if !strings.Contains(err.Error(), key) {
				t.Errorf("error should mention the unknown key: %v", err)
			}
		})
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

	cfg, err := config.LoadConfigFrom(dir)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if cfg.ConfigHome != dir {
		t.Errorf("ConfigHome: got %q, want %q", cfg.ConfigHome, dir)
	}

	if len(cfg.AdditionalMounts) != 0 {
		t.Errorf("AdditionalMounts: got %v, want empty", cfg.AdditionalMounts)
	}

	if len(cfg.InheritEnv) != 0 {
		t.Errorf("InheritEnv: got %v, want empty", cfg.InheritEnv)
	}

	if len(cfg.EnvFiles) != 0 {
		t.Errorf("EnvFiles: got %v, want empty", cfg.EnvFiles)
	}

	if len(cfg.RuntimeOptions) != 0 {
		t.Errorf("RuntimeOptions: got %v, want empty", cfg.RuntimeOptions)
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
inherit_env = ["SSH_AUTH_SOCK"]
env_files = [".env"]

[[runtime_options]]
runtime = "podman"
subcommand = "run"
args = ["--userns=keep-id"]
`)

	cfg, err := config.LoadConfigFrom(dir)
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
	assertStringSlice(t, "InheritEnv", cfg.InheritEnv, []string{testInheritEnv})
	assertStringSlice(t, "EnvFiles", cfg.EnvFiles, []string{testEnvFile})

	wantRuntimeOptions := []config.RuntimeOption{
		{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
	}
	if !reflect.DeepEqual(cfg.RuntimeOptions, wantRuntimeOptions) {
		t.Errorf("RuntimeOptions: got %v, want %v", cfg.RuntimeOptions, wantRuntimeOptions)
	}
}

func TestLoadConfig_XDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	chellyDir := filepath.Join(dir, "chelly")
	writeConfigFile(t, chellyDir, `
container_cmd = "podman"
`)

	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := config.LoadConfig()
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
inherit_env = ["CONFIG_FILE_TOKEN"]
env_files = ["config-file.env"]
`)

	t.Setenv("CHELLY_CONTAINER_CMD", testContainerCmdPodman)
	t.Setenv("CHELLY_CONFIG_HOME", "/env-context")
	t.Setenv("CHELLY_WORKDIR", "/env-workdir")
	t.Setenv("CHELLY_ADDITIONAL_MOUNTS", "/env-host:/env-container")
	t.Setenv("CHELLY_INHERIT_ENV", testInheritEnv+","+testInheritEnv2)
	t.Setenv("CHELLY_ENV_FILES", testEnvFile+","+testEnvFile2)

	cfg, err := config.LoadConfigFrom(dir)
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
	assertStringSlice(t, "InheritEnv", cfg.InheritEnv, []string{testInheritEnv, testInheritEnv2})
	assertStringSlice(t, "EnvFiles", cfg.EnvFiles, []string{testEnvFile, testEnvFile2})
}

func TestValidateInheritEnv(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		values  []string
		wantErr bool
	}{
		{testCaseNameEmpty, nil, false},
		{"valid", []string{"SSH_AUTH_SOCK", "_TOKEN", "GITHUB_TOKEN2"}, false},
		{"contains equals", []string{"FOO=bar"}, true},
		{"starts with number", []string{"1TOKEN"}, true},
		{"contains hyphen", []string{"GITHUB-TOKEN"}, true},
		{"empty name", []string{""}, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateInheritEnv(testCase.values)
			if testCase.wantErr && err == nil {
				t.Fatal("ValidateInheritEnv returned nil, want error")
			}

			if !testCase.wantErr && err != nil {
				t.Fatalf("ValidateInheritEnv: %v", err)
			}
		})
	}
}

func TestValidateRuntimeOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		opts    []config.RuntimeOption
		wantErr bool
	}{
		{testCaseNameEmpty, nil, false},
		{
			"valid",
			[]config.RuntimeOption{
				{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
				{Runtime: testRuntimeDocker, Subcommand: config.SubcommandBuild, Args: []string{testExtraArg2}},
			},
			false,
		},
		{
			"invalid subcommand",
			[]config.RuntimeOption{{Runtime: testRuntimePodman, Subcommand: "exec", Args: []string{testExtraArg}}},
			true,
		},
		{
			"duplicate runtime and subcommand",
			[]config.RuntimeOption{
				{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
				{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg2}},
			},
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateRuntimeOptions(testCase.opts)
			if testCase.wantErr && err == nil {
				t.Fatal("ValidateRuntimeOptions returned nil, want error")
			}

			if !testCase.wantErr && err != nil {
				t.Fatalf("ValidateRuntimeOptions: %v", err)
			}
		})
	}
}

func TestResolveRuntimeArgs(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		ContainerCmd:     testContainerCmdPodman,
		ConfigHome:       "",
		Workdir:          "",
		AdditionalMounts: nil,
		InheritEnv:       nil,
		EnvFiles:         nil,
		RuntimeOptions: []config.RuntimeOption{
			{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
		},
	}

	assertStringSlice(t, "run", config.ResolveRuntimeArgs(cfg, config.SubcommandRun), []string{testExtraArg})
	assertStringSlice(t, "build", config.ResolveRuntimeArgs(cfg, config.SubcommandBuild), nil)
}

func TestResolveRuntimeArgs_UsesContainerCmdBasename(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		ContainerCmd:     "/usr/bin/podman",
		ConfigHome:       "",
		Workdir:          "",
		AdditionalMounts: nil,
		InheritEnv:       nil,
		EnvFiles:         nil,
		RuntimeOptions: []config.RuntimeOption{
			{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
		},
	}

	assertStringSlice(t, "run", config.ResolveRuntimeArgs(cfg, config.SubcommandRun), []string{testExtraArg})
}

func TestResolveRuntimeArgs_IgnoresOtherRuntime(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		ContainerCmd:     testContainerCmdDocker,
		ConfigHome:       "",
		Workdir:          "",
		AdditionalMounts: nil,
		InheritEnv:       nil,
		EnvFiles:         nil,
		RuntimeOptions: []config.RuntimeOption{
			{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
		},
	}

	assertStringSlice(t, "run", config.ResolveRuntimeArgs(cfg, config.SubcommandRun), nil)
}

func TestResolveRuntimeArgs_EnvVarOverridesConfig(t *testing.T) {
	cfg := config.Config{
		ContainerCmd:     testContainerCmdPodman,
		ConfigHome:       "",
		Workdir:          "",
		AdditionalMounts: nil,
		InheritEnv:       nil,
		EnvFiles:         nil,
		RuntimeOptions: []config.RuntimeOption{
			{Runtime: testRuntimePodman, Subcommand: config.SubcommandRun, Args: []string{testExtraArg}},
		},
	}

	t.Setenv("CHELLY_RUNTIME_OPTIONS_PODMAN_RUN", testExtraArg2)

	assertStringSlice(t, "run", config.ResolveRuntimeArgs(cfg, config.SubcommandRun), []string{testExtraArg2})
}

func writeEnvFile(t *testing.T, dir, name string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("FOO=bar\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	return path
}

func TestResolveEnvFiles_AbsoluteAndRelative(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	absFile := writeEnvFile(t, dir, "abs.env")
	writeEnvFile(t, dir, "rel.env")

	resolved, skipped, err := config.ResolveEnvFiles([]string{absFile, "rel.env"}, dir)
	if err != nil {
		t.Fatalf("ResolveEnvFiles: %v", err)
	}

	assertStringSlice(t, "resolved", resolved, []string{absFile, filepath.Join(dir, "rel.env")})
	assertStringSlice(t, "skipped", skipped, nil)
}

func TestResolveEnvFiles_MissingRelativeIsSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeEnvFile(t, dir, "present.env")

	resolved, skipped, err := config.ResolveEnvFiles([]string{"missing.env", "present.env"}, dir)
	if err != nil {
		t.Fatalf("ResolveEnvFiles: %v", err)
	}

	assertStringSlice(t, "resolved", resolved, []string{filepath.Join(dir, "present.env")})
	assertStringSlice(t, "skipped", skipped, []string{"missing.env"})
}

func TestResolveEnvFiles_MissingAbsoluteIsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	_, _, err := config.ResolveEnvFiles([]string{filepath.Join(dir, "missing.env")}, dir)
	if !errors.Is(err, config.ErrEnvFileNotFound) {
		t.Fatalf("got %v, want ErrEnvFileNotFound", err)
	}
}

func TestResolveEnvFiles_TildeExpansion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeEnvFile(t, home, "home.env")

	resolved, skipped, err := config.ResolveEnvFiles([]string{"~/home.env"}, t.TempDir())
	if err != nil {
		t.Fatalf("ResolveEnvFiles: %v", err)
	}

	assertStringSlice(t, "resolved", resolved, []string{filepath.Join(home, "home.env")})
	assertStringSlice(t, "skipped", skipped, nil)
}

func TestResolveEnvFiles_MissingTildePathIsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, _, err := config.ResolveEnvFiles([]string{"~/missing.env"}, t.TempDir())
	if !errors.Is(err, config.ErrEnvFileNotFound) {
		t.Fatalf("got %v, want ErrEnvFileNotFound", err)
	}
}
