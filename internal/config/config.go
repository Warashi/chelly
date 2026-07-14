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

// Package config manages chelly configuration loading, formatting, and updates.
package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

const (
	dirPerm             = 0o700
	filePerm            = 0o600
	keyAdditionalMounts = "additional_mounts"
	keyInheritEnv       = "inherit_env"
	keyEnvFiles         = "env_files"
	keyRuntimeOptions   = "runtime_options"
)

const (
	// SubcommandRun is the container runtime subcommand used to run a container.
	SubcommandRun = "run"
	// SubcommandBuild is the container runtime subcommand used to build an image.
	SubcommandBuild = "build"
)

// ErrUnknownConfigKey is returned when an unrecognized configuration key is used.
var ErrUnknownConfigKey = errors.New("unknown config key")

// ErrInvalidEnvName is returned when inherit_env contains an invalid environment variable name.
var ErrInvalidEnvName = errors.New("invalid environment variable name")

// ErrEnvFileNotFound is returned when an env_files entry with an absolute path does not exist.
var ErrEnvFileNotFound = errors.New("env file not found")

// ErrInvalidSubcommand is returned when a runtime_options entry names an unsupported subcommand.
var ErrInvalidSubcommand = errors.New("invalid runtime_options subcommand")

// ErrDuplicateRuntimeOption is returned when runtime_options contains more than one
// entry for the same runtime and subcommand pair.
var ErrDuplicateRuntimeOption = errors.New("duplicate runtime_options entry")

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// runtimeOptionKeyPattern matches keys of the form runtime_options.<runtime>.<subcommand>.
var runtimeOptionKeyPattern = regexp.MustCompile(`^runtime_options\.([^.]+)\.([^.]+)$`)

// validSubcommands is the list of container runtime subcommands that support
// runtime-specific extra arguments.
var validSubcommands = []string{SubcommandRun, SubcommandBuild}

// RuntimeOption holds extra arguments to pass to a specific container runtime
// subcommand (e.g. "podman run" or "docker build").
type RuntimeOption struct {
	Runtime    string   `toml:"runtime"`
	Subcommand string   `toml:"subcommand"`
	Args       []string `toml:"args"`
}

// Config holds the chelly configuration.
type Config struct {
	ContainerCmd     string          `toml:"container_cmd"`
	ConfigHome       string          `toml:"config_home"`
	Workdir          string          `toml:"workdir"`
	AdditionalMounts []string        `toml:"additional_mounts"`
	InheritEnv       []string        `toml:"inherit_env"`
	EnvFiles         []string        `toml:"env_files"`
	RuntimeOptions   []RuntimeOption `toml:"runtime_options"`
}

// validConfigKeys is the list of all valid static configuration key names.
// Dynamic keys of the form runtime_options.<runtime>.<subcommand> are validated separately.
var validConfigKeys = []string{
	"container_cmd",
	"config_home",
	"workdir",
	keyAdditionalMounts,
	keyInheritEnv,
	keyEnvFiles,
}

// FormatConfig serializes cfg to a TOML string representing the effective configuration.
func FormatConfig(cfg Config) (string, error) {
	b, err := toml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshaling config: %w", err)
	}

	return string(b), nil
}

// GetConfigValue returns the effective value of the named key as a string.
// For list-valued settings, values are comma-joined.
func GetConfigValue(cfg Config, key string) (string, error) {
	if runtime, subcommand, ok := parseRuntimeOptionKey(key); ok {
		return getRuntimeOptionValue(cfg, runtime, subcommand)
	}

	switch key {
	case "container_cmd":
		return cfg.ContainerCmd, nil
	case "config_home":
		return cfg.ConfigHome, nil
	case "workdir":
		return cfg.Workdir, nil
	case keyAdditionalMounts:
		return strings.Join(cfg.AdditionalMounts, ","), nil
	case keyInheritEnv:
		return strings.Join(cfg.InheritEnv, ","), nil
	case keyEnvFiles:
		return strings.Join(cfg.EnvFiles, ","), nil
	default:
		return "", fmt.Errorf("%w %q: valid keys are %s", ErrUnknownConfigKey, key, strings.Join(validConfigKeys, ", "))
	}
}

func getRuntimeOptionValue(cfg Config, runtime, subcommand string) (string, error) {
	if err := validateSubcommand(subcommand); err != nil {
		return "", err
	}

	for _, opt := range cfg.RuntimeOptions {
		if opt.Runtime == runtime && opt.Subcommand == subcommand {
			return strings.Join(opt.Args, ","), nil
		}
	}

	return "", nil
}

// parseRuntimeOptionKey splits a key of the form runtime_options.<runtime>.<subcommand>
// into its runtime and subcommand parts. ok is false when key does not match this shape.
func parseRuntimeOptionKey(key string) (string, string, bool) {
	m := runtimeOptionKeyPattern.FindStringSubmatch(key)
	if m == nil {
		return "", "", false
	}

	return m[1], m[2], true
}

func validateSubcommand(subcommand string) error {
	if !slices.Contains(validSubcommands, subcommand) {
		return fmt.Errorf("%w %q: valid subcommands are %s",
			ErrInvalidSubcommand, subcommand, strings.Join(validSubcommands, ", "))
	}

	return nil
}

func configListValue(value string) []string {
	var items []string

	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}

	return items
}

func configListValueFrom(viperInst *viper.Viper, key string) []string {
	envName := "CHELLY_" + strings.ToUpper(strings.NewReplacer(".", "_").Replace(key))
	if value, ok := os.LookupEnv(envName); ok {
		return configListValue(value)
	}

	return viperInst.GetStringSlice(key)
}

func applyConfigValue(data map[string]any, key, value string) {
	var configValue any

	switch key {
	case keyAdditionalMounts, keyInheritEnv, keyEnvFiles:
		configValue = configListValue(value)
	default:
		configValue = value
	}

	parts := strings.Split(key, ".")

	current := data
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}

		current = next
	}

	current[parts[len(parts)-1]] = configValue
}

// applyRuntimeOptionValue upserts the entry matching runtime and subcommand within
// data[keyRuntimeOptions] (an array of tables), setting its args to value's
// comma-separated list.
func applyRuntimeOptionValue(data map[string]any, runtime, subcommand, value string) {
	entries, _ := data[keyRuntimeOptions].([]any)

	args := configListValue(value)

	for _, entry := range entries {
		table, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		if table["runtime"] == runtime && table["subcommand"] == subcommand {
			table["args"] = args

			data[keyRuntimeOptions] = entries

			return
		}
	}

	entries = append(entries, map[string]any{
		"runtime":    runtime,
		"subcommand": subcommand,
		"args":       args,
	})

	data[keyRuntimeOptions] = entries
}

// SetConfigValue writes key=value into the TOML config file in configDir.
// For additional_mounts, value is a comma-separated list of mount specs.
// The config file and directory are created if they do not exist.
func SetConfigValue(configDir, key, value string) error {
	runtime, subcommand, isRuntimeOption := parseRuntimeOptionKey(key)
	if isRuntimeOption {
		if err := validateSubcommand(subcommand); err != nil {
			return err
		}
	} else if !slices.Contains(validConfigKeys, key) {
		return fmt.Errorf("%w %q: valid keys are %s", ErrUnknownConfigKey, key, strings.Join(validConfigKeys, ", "))
	}

	data, err := readConfigData(configDir)
	if err != nil {
		return err
	}

	if isRuntimeOption {
		applyRuntimeOptionValue(data, runtime, subcommand, value)
	} else {
		applyConfigValue(data, key, value)
	}

	return writeConfigData(configDir, data)
}

func readConfigData(configDir string) (map[string]any, error) {
	configFile := filepath.Join(configDir, "config.toml")

	data := map[string]any{}

	content, err := os.ReadFile(configFile) //nolint:gosec
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data, nil
		}

		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := toml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return data, nil
}

func writeConfigData(configDir string, data map[string]any) error {
	content, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(configDir, dirPerm); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	configFile := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configFile, content, filePerm); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// ValidateInheritEnv returns an error when inherit_env contains an invalid environment variable name.
func ValidateInheritEnv(names []string) error {
	for _, name := range names {
		if !envNamePattern.MatchString(name) {
			return fmt.Errorf("%w %q", ErrInvalidEnvName, name)
		}
	}

	return nil
}

// ResolveEnvFiles resolves env_files entries into absolute paths to pass to the
// container runtime.
//
// Entries starting with "~" are expanded to the user home directory and then
// treated as absolute paths. Absolute paths that do not exist cause an error.
// Relative paths are resolved against baseDir; missing ones are skipped and
// returned in the second slice so the caller can warn about them.
func ResolveEnvFiles(entries []string, baseDir string) ([]string, []string, error) {
	var resolved, skipped []string

	for _, entry := range entries {
		path, wasRelative, err := expandEnvFilePath(entry, baseDir)
		if err != nil {
			return nil, nil, err
		}

		if _, err := os.Stat(path); err != nil {
			if wasRelative && errors.Is(err, os.ErrNotExist) {
				skipped = append(skipped, entry)

				continue
			}

			return nil, nil, fmt.Errorf("%w %q: %w", ErrEnvFileNotFound, entry, err)
		}

		resolved = append(resolved, path)
	}

	return resolved, skipped, nil
}

func expandEnvFilePath(entry, baseDir string) (string, bool, error) {
	if entry == "~" || strings.HasPrefix(entry, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false, fmt.Errorf("expanding %q: %w", entry, err)
		}

		return filepath.Join(home, strings.TrimPrefix(entry, "~")), false, nil
	}

	if filepath.IsAbs(entry) {
		return entry, false, nil
	}

	return filepath.Join(baseDir, entry), true, nil
}

// ValidateRuntimeOptions returns an error when opts contains an entry with an
// unsupported subcommand, or more than one entry for the same runtime and
// subcommand pair.
func ValidateRuntimeOptions(opts []RuntimeOption) error {
	seen := map[[2]string]struct{}{}

	for _, opt := range opts {
		if err := validateSubcommand(opt.Subcommand); err != nil {
			return err
		}

		pair := [2]string{opt.Runtime, opt.Subcommand}
		if _, ok := seen[pair]; ok {
			return fmt.Errorf("%w: runtime %q, subcommand %q", ErrDuplicateRuntimeOption, opt.Runtime, opt.Subcommand)
		}

		seen[pair] = struct{}{}
	}

	return nil
}

// ResolveRuntimeArgs returns the extra arguments to pass to the given container
// runtime subcommand (e.g. SubcommandRun or SubcommandBuild), based on cfg.ContainerCmd.
//
// The environment variable CHELLY_RUNTIME_OPTIONS_<RUNTIME>_<SUBCOMMAND> takes
// precedence over cfg.RuntimeOptions when set.
func ResolveRuntimeArgs(cfg Config, subcommand string) []string {
	runtime := filepath.Base(cfg.ContainerCmd)

	envName := "CHELLY_RUNTIME_OPTIONS_" + strings.ToUpper(runtime) + "_" + strings.ToUpper(subcommand)
	if value, ok := os.LookupEnv(envName); ok {
		return configListValue(value)
	}

	for _, opt := range cfg.RuntimeOptions {
		if opt.Runtime == runtime && opt.Subcommand == subcommand {
			return opt.Args
		}
	}

	return nil
}

// DetectContainerCmd returns the first available container runtime found in PATH.
func DetectContainerCmd() string {
	for _, name := range []string{"container", "podman", "docker"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}

	return "docker"
}

// DefaultConfigDir returns the chelly configuration directory.
func DefaultConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "chelly"), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting user config dir: %w", err)
	}

	return filepath.Join(dir, "chelly"), nil
}

// LoadConfig loads configuration from the default location and environment variables.
func LoadConfig() (Config, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return Config{}, err
	}

	return LoadConfigFrom(dir)
}

// LoadConfigFrom loads configuration from the given configDir and environment variables.
func LoadConfigFrom(configDir string) (Config, error) {
	viperInst := viper.New()

	viperInst.SetConfigName("config")
	viperInst.SetConfigType("toml")
	viperInst.AddConfigPath(configDir)

	viperInst.SetEnvPrefix("CHELLY")
	viperInst.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperInst.AutomaticEnv()

	viperInst.SetDefault("container_cmd", DetectContainerCmd())
	viperInst.SetDefault("config_home", configDir)
	viperInst.SetDefault("additional_mounts", []string{})
	viperInst.SetDefault("inherit_env", []string{})
	viperInst.SetDefault("env_files", []string{})

	if err := viperInst.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("reading config file: %w", err)
		}
	}

	var runtimeOptions []RuntimeOption
	if err := viperInst.UnmarshalKey(keyRuntimeOptions, &runtimeOptions); err != nil {
		return Config{}, fmt.Errorf("parsing runtime_options: %w", err)
	}

	workdir := viperInst.GetString("workdir")
	if workdir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return Config{}, fmt.Errorf("getting working directory: %w", err)
		}

		workdir = currentDir
	}

	return Config{
		ContainerCmd:     viperInst.GetString("container_cmd"),
		ConfigHome:       viperInst.GetString("config_home"),
		Workdir:          workdir,
		AdditionalMounts: configListValueFrom(viperInst, keyAdditionalMounts),
		InheritEnv:       configListValueFrom(viperInst, keyInheritEnv),
		EnvFiles:         configListValueFrom(viperInst, keyEnvFiles),
		RuntimeOptions:   runtimeOptions,
	}, nil
}
