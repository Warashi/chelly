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

package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

const (
	dirPerm               = 0o700
	filePerm              = 0o600
	keyAdditionalMounts   = "additional_mounts"
	keyContainerSetupCmds = "container_setup_cmds"
)

// ErrUnknownConfigKey is returned when an unrecognized configuration key is used.
var ErrUnknownConfigKey = errors.New("unknown config key")

// Config holds the chelly configuration.
type Config struct {
	ContainerCmd       string   `toml:"container_cmd"`
	ConfigHome         string   `toml:"config_home"`
	Workdir            string   `toml:"workdir"`
	AdditionalMounts   []string `toml:"additional_mounts"`
	ContainerSetupCmds []string `toml:"container_setup_cmds"`
}

// validConfigKeys is the list of all valid configuration key names.
var validConfigKeys = []string{
	"container_cmd",
	"config_home",
	"workdir",
	keyAdditionalMounts,
	keyContainerSetupCmds,
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
// For additional_mounts and container_setup_cmds, values are comma-joined.
func GetConfigValue(cfg Config, key string) (string, error) {
	switch key {
	case "container_cmd":
		return cfg.ContainerCmd, nil
	case "config_home":
		return cfg.ConfigHome, nil
	case "workdir":
		return cfg.Workdir, nil
	case keyAdditionalMounts:
		return strings.Join(cfg.AdditionalMounts, ","), nil
	case keyContainerSetupCmds:
		return strings.Join(cfg.ContainerSetupCmds, ","), nil
	default:
		return "", fmt.Errorf("%w %q: valid keys are %s", ErrUnknownConfigKey, key, strings.Join(validConfigKeys, ", "))
	}
}

func applyConfigValue(data map[string]any, key, value string) {
	switch key {
	case keyAdditionalMounts, keyContainerSetupCmds:
		var items []string

		for item := range strings.SplitSeq(value, ",") {
			if item = strings.TrimSpace(item); item != "" {
				items = append(items, item)
			}
		}

		data[key] = items
	default:
		data[key] = value
	}
}

// SetConfigValue writes key=value into the TOML config file in configDir.
// For additional_mounts, value is a comma-separated list of mount specs.
// The config file and directory are created if they do not exist.
func SetConfigValue(configDir, key, value string) error {
	if !slices.Contains(validConfigKeys, key) {
		return fmt.Errorf("%w %q: valid keys are %s", ErrUnknownConfigKey, key, strings.Join(validConfigKeys, ", "))
	}

	configFile := filepath.Join(configDir, "config.toml")

	data := map[string]any{}

	if content, err := os.ReadFile(configFile); err == nil { //nolint:gosec
		if err := toml.Unmarshal(content, &data); err != nil {
			return fmt.Errorf("parsing config file: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading config file: %w", err)
	}

	applyConfigValue(data, key, value)

	content, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(configDir, dirPerm); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	if err := os.WriteFile(configFile, content, filePerm); err != nil {
		return fmt.Errorf("writing config file: %w", err)
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

func chellyConfigDir() (string, error) {
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
	dir, err := chellyConfigDir()
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
	viperInst.AutomaticEnv()

	viperInst.SetDefault("container_cmd", DetectContainerCmd())
	viperInst.SetDefault("config_home", configDir)
	viperInst.SetDefault("additional_mounts", []string{})
	viperInst.SetDefault("container_setup_cmds", []string{})

	if err := viperInst.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("reading config file: %w", err)
		}
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
		ContainerCmd:       viperInst.GetString("container_cmd"),
		ConfigHome:         viperInst.GetString("config_home"),
		Workdir:            workdir,
		AdditionalMounts:   viperInst.GetStringSlice("additional_mounts"),
		ContainerSetupCmds: viperInst.GetStringSlice("container_setup_cmds"),
	}, nil
}
