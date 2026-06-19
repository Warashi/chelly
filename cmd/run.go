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
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:                "run [-- command [args...]]",
	Short:              "Run a command in the chelly container",
	Long:               `Run a command inside the chelly container, mounting the current directory.`,
	DisableFlagParsing: true,
	RunE:               runRun,
}

var runitCmd = &cobra.Command{
	Use:   "runit [-- command [args...]]",
	Short: "Start an interactive shell in the chelly container",
	Long: `Start an interactive shell in the chelly container.
Always forces --interactive --tty. Defaults to "sh" if no command is given.`,
	DisableFlagParsing: true,
	RunE:               runRunit,
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(runitCmd)
}

// IsTTY reports whether f is connected to a terminal.
func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

// ContainerRunArgs returns the argument slice for the container run command.
func ContainerRunArgs(cfg Config, workDir string, forceTTY bool, isTTY bool, userArgs []string) []string {
	args := []string{"run", "--rm"}

	if forceTTY || isTTY {
		args = append(args, "--interactive", "--tty")
	}

	args = append(args, "--volume", "chelly-home:/home")
	args = append(args, "--volume", workDir+":"+workDir)

	for _, mount := range cfg.AdditionalMounts {
		args = append(args, "--volume", mount)
	}

	args = append(args, "--workdir", cfg.Workdir)
	args = append(args, "chelly:latest")

	if cfg.ContainerSetupCmd != "" {
		if len(userArgs) == 0 {
			args = append(args, "sh", "-lc", cfg.ContainerSetupCmd)
		} else {
			args = append(args, "sh", "-lc", cfg.ContainerSetupCmd+`; exec "$@"`, "sh")
			args = append(args, userArgs...)
		}
	} else {
		args = append(args, userArgs...)
	}

	return args
}

func stripDashDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}

	return args
}

func execContainer(ctx context.Context, containerCmd string, args []string) error {
	cmd := exec.CommandContext(ctx, containerCmd, args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running container: %w", err)
	}

	return nil
}

func runRun(cobraCmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	tty := IsTTY(os.Stdin) && IsTTY(os.Stdout)
	userArgs := stripDashDash(args)
	containerArgs := ContainerRunArgs(cfg, currentDir, false, tty, userArgs)

	return execContainer(cobraCmd.Context(), cfg.ContainerCmd, containerArgs)
}

func runRunit(cobraCmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	userArgs := stripDashDash(args)
	if len(userArgs) == 0 {
		userArgs = []string{"sh"}
	}

	containerArgs := ContainerRunArgs(cfg, currentDir, true, false, userArgs)

	return execContainer(cobraCmd.Context(), cfg.ContainerCmd, containerArgs)
}
