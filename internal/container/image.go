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

package container

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrPullOptionConflict is returned when caller-supplied run arguments contain a
// --pull option, which would override chelly's local-image-only policy.
var ErrPullOptionConflict = errors.New("--pull is managed by chelly and cannot be set in runtime_options")

const (
	flagPull      = "--pull"
	pullNeverFlag = "--pull=never"
)

// runtimePolicy describes how a container runtime keeps `chelly run` on the
// locally built image.
type runtimePolicy struct {
	// pullArgs are the run options that forbid pulling; empty when the runtime
	// offers no such option (Apple container 1.4.1 `run` has none).
	pullArgs []string
}

// runtimePolicies is keyed by the basename of the container command.
var runtimePolicies = map[string]runtimePolicy{
	"podman":    {pullArgs: []string{pullNeverFlag}},
	"docker":    {pullArgs: []string{pullNeverFlag}},
	"container": {pullArgs: nil},
}

// RuntimeName returns the runtime identity used to look up runtime-specific behavior.
func RuntimeName(containerCmd string) string {
	return filepath.Base(containerCmd)
}

// pullPolicyArgs returns the run options that stop the runtime from pulling the image.
func pullPolicyArgs(containerCmd string) []string {
	return runtimePolicies[RuntimeName(containerCmd)].pullArgs
}

// ValidateRunExtraArgs rejects run arguments that would change the pull policy.
func ValidateRunExtraArgs(extraArgs []string) error {
	for _, arg := range extraArgs {
		if arg == flagPull || strings.HasPrefix(arg, flagPull+"=") {
			return fmt.Errorf("%w: %q", ErrPullOptionConflict, arg)
		}
	}

	return nil
}
