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
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var noCache bool

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the chelly container image",
	Long:  `Build the chelly container image using the context directory in $XDG_CONFIG_HOME/chelly.`,
	RunE:  runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "Do not use cache when building the image")
}

// BuildArgs returns the argument slice for the container build command.
func BuildArgs(cfg Config, noCacheFlag bool) []string {
	args := []string{"build"}
	if noCacheFlag {
		args = append(args, "--no-cache")
	}

	args = append(args, "--tag", "chelly:latest", cfg.ConfigHome)

	return args
}

func runBuild(cobraCmd *cobra.Command, _ []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	args := BuildArgs(cfg, noCache)

	cmd := exec.CommandContext(cobraCmd.Context(), cfg.ContainerCmd, args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running build command: %w", err)
	}

	return nil
}
