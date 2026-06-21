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

// Package cmd defines chelly's Cobra command tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// NewRootCommand returns the root chelly command with all subcommands registered.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "chelly",
		Short: "A personal container-based development environment manager",
		Long: `chelly manages a container image built from your ~/.config/chelly directory
and provides commands to run tools inside that container.`,
	}

	rootCmd.AddCommand(newBuildCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newConfigCommand())

	return rootCmd
}

// Execute runs the root command.
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
