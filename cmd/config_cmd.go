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
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage chelly configuration",
	Long:  `Manage chelly configuration. Use subcommands to list, get, or set configuration values.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all current effective configuration values",
	Long: `Print all effective configuration values in TOML format,
reflecting config file and environment variable overrides.`,
	RunE: runConfigList,
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a specific configuration value",
	Long: `Print the effective value of a single configuration key,
reflecting config file and environment variable overrides.`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

const configSetArgs = 2

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value in config.toml",
	Long: `Write a configuration key-value pair to the config.toml file.
For list values, provide a comma-separated list.`,
	Args: cobra.ExactArgs(configSetArgs),
	RunE: runConfigSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
}

func runConfigList(cobraCmd *cobra.Command, _ []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	out, err := FormatConfig(cfg)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(cobraCmd.OutOrStdout(), out); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func runConfigGet(cobraCmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	val, err := GetConfigValue(cfg, args[0])
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(cobraCmd.OutOrStdout(), val); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func runConfigSet(_ *cobra.Command, args []string) error {
	configDir, err := chellyConfigDir()
	if err != nil {
		return err
	}

	return SetConfigValue(configDir, args[0], args[1])
}
