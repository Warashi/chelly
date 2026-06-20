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

const (
	testContainerCmdDocker = "docker"
	testContainerCmdPodman = "podman"
	testConfigHome         = "/config/chelly"
	testWorkDir            = "/current/dir"
	testWorkDirMount       = "/current/dir:/current/dir"
	testWorkspace          = "/workspace"
	testSetupCmd           = "echo setup"
	testSetupCmd2          = "echo setup2"
	testMountA             = "/a:/a"
	testMountB             = "/b:/b"

	imageChelly     = "chelly:latest"
	flagRM          = "--rm"
	flagVolume      = "--volume"
	flagWorkdir     = "--workdir"
	flagInteractive = "--interactive"
	flagTTY         = "--tty"
	shellSh         = "sh"
	shellFlagLC     = "-lc"
	cmdRun          = "run"
	cmdEcho         = "echo"
	cmdBash         = "bash"
	cmdHello        = "hello"
)
