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

package cmd_test

import (
	"reflect"
	"testing"

	"github.com/Warashi/chelly/cmd"
)

func baseConfig() cmd.Config {
	return cmd.Config{
		ContainerCmd:       testContainerCmdDocker,
		ConfigHome:         testConfigHome,
		Workdir:            testWorkspace,
		AdditionalMounts:   nil,
		ContainerSetupCmds: nil,
	}
}

func TestBuildArgs_WithoutNoCache(t *testing.T) {
	t.Parallel()

	cfg := baseConfig()
	got := cmd.BuildArgs(cfg, false)
	want := []string{"build", "--tag", imageChelly, testConfigHome}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildArgs: got %v, want %v", got, want)
	}
}

func TestBuildArgs_WithNoCache(t *testing.T) {
	t.Parallel()

	cfg := baseConfig()
	got := cmd.BuildArgs(cfg, true)
	want := []string{"build", "--no-cache", "--tag", imageChelly, testConfigHome}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildArgs: got %v, want %v", got, want)
	}
}
