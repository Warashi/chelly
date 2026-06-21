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

	"github.com/Warashi/chelly/internal/config"
	"github.com/Warashi/chelly/internal/container"
	"github.com/spf13/cobra"
)

func newBuildCommand() *cobra.Command {
	var noCache bool

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build the chelly container image",
		Long:  `Build the chelly container image using the context directory in $XDG_CONFIG_HOME/chelly.`,
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			args := container.BuildArgs(container.BuildConfig{ConfigHome: cfg.ConfigHome}, noCache)

			return container.Run(cobraCmd.Context(), cfg.ContainerCmd, args)
		},
	}

	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "Do not use cache when building the image")

	return buildCmd
}
