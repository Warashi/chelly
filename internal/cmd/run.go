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

	"github.com/Warashi/chelly/internal/config"
	"github.com/Warashi/chelly/internal/container"
	"github.com/Warashi/chelly/internal/git"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "run [-- command [args...]]",
		Short:              "Run a command in the chelly container",
		Long:               `Run a command inside the chelly container, mounting the current directory.`,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err := config.ValidateInheritEnv(cfg.InheritEnv); err != nil {
				return fmt.Errorf("validating inherit_env: %w", err)
			}

			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			autoMounts := []string{}
			if commonParent, ok := git.LinkedWorktreeCommonParent(currentDir); ok {
				autoMounts = append(autoMounts, commonParent)
			}

			tty := container.IsTTY(os.Stdin) && container.IsTTY(os.Stdout)
			userArgs := container.StripDashDash(args)
			containerArgs := container.RunArgs(container.RunConfig{
				ContainerCmd:       cfg.ContainerCmd,
				Workdir:            cfg.Workdir,
				AdditionalMounts:   cfg.AdditionalMounts,
				ContainerSetupCmds: cfg.ContainerSetupCmds,
				InheritEnv:         cfg.InheritEnv,
				PodmanOptions:      container.PodmanOptions{Run: cfg.PodmanOptions.Run},
			}, currentDir, tty, userArgs, autoMounts...)

			return container.Exec(cfg.ContainerCmd, containerArgs)
		},
	}
}
