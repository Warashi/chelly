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

package container_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Warashi/chelly/internal/container"
)

const testContainerCmdApple = "container"

// Apple container 1.4.1 `run` has no pull policy option, so chelly must not
// emit one there.
func TestRunArgs_AppleContainerHasNoPullFlag(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerCmd = testContainerCmdApple

	got := container.RunArgs(cfg, testWorkDir, false, []string{"ls"})
	want := []string{
		cmdRun, flagRM, flagInteractive,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
		"ls",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestRunArgs_PullNeverWithPathContainerCmd(t *testing.T) {
	t.Parallel()

	cfg := baseRunConfig()
	cfg.ContainerCmd = "/usr/bin/podman"

	got := container.RunArgs(cfg, testWorkDir, false, nil)
	want := []string{
		cmdRun, flagRM, flagInteractive, flagPullNever,
		flagVolume, testWorkDirMount,
		flagWorkdir, testWorkDir,
		container.ImageName,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RunArgs: got %v, want %v", got, want)
	}
}

func TestValidateRunExtraArgs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args    []string
		wantErr error
	}{
		"no pull option":      {args: []string{testExtraArg, "--pull-secret", "x"}, wantErr: nil},
		"pull with equals":    {args: []string{"--pull=always"}, wantErr: container.ErrPullOptionConflict},
		"pull separate value": {args: []string{flagRM, "--pull", "never"}, wantErr: container.ErrPullOptionConflict},
		"empty":               {args: nil, wantErr: nil},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := container.ValidateRunExtraArgs(testCase.args)
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("ValidateRunExtraArgs(%v) = %v, want %v", testCase.args, err, testCase.wantErr)
			}
		})
	}
}
