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

	"github.com/spf13/viper"
)

// Config holds the chelly configuration.
type Config struct {
	ContainerCmd      string
	ConfigHome        string
	Workdir           string
	AdditionalMounts  []string
	ContainerSetupCmd string
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
	viperInst.SetDefault("container_setup_cmd", "")

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
		ContainerCmd:      viperInst.GetString("container_cmd"),
		ConfigHome:        viperInst.GetString("config_home"),
		Workdir:           workdir,
		AdditionalMounts:  viperInst.GetStringSlice("additional_mounts"),
		ContainerSetupCmd: viperInst.GetString("container_setup_cmd"),
	}, nil
}
