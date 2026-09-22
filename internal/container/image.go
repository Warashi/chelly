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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// ErrPullOptionConflict is returned when caller-supplied run arguments contain a
	// --pull option, which would override chelly's local-image-only policy.
	ErrPullOptionConflict = errors.New("--pull is managed by chelly and cannot be set in runtime_options")

	// ErrImageNotFound is returned when the chelly image does not exist locally.
	ErrImageNotFound = errors.New("chelly image not found locally")

	// ErrUnsupportedRuntime is returned when chelly cannot guarantee local-image-only
	// execution for the configured container command.
	ErrUnsupportedRuntime = errors.New("unsupported container runtime for chelly run")
)

const (
	flagPull      = "--pull"
	pullNeverFlag = "--pull=never"
)

// imageCheck is how a runtime answers "does ImageName exist locally?".
type imageCheck int

const (
	// checkByList runs `image ls -q IMAGE`: exit 0 with empty output means
	// missing, so connection and permission failures stay distinguishable.
	checkByList imageCheck = iota
	// checkByInspect runs `image inspect IMAGE`: any failure is reported as
	// missing with the runtime's stderr passed through, because Apple container
	// 1.4.1 `image list` cannot filter by reference.
	checkByInspect
)

// runtimePolicy describes how a container runtime keeps `chelly run` on the
// locally built image.
type runtimePolicy struct {
	// pullArgs are the run options that forbid pulling; empty when the runtime
	// offers no such option (Apple container 1.4.1 `run` has none).
	pullArgs []string
	check    imageCheck
}

// runtimePolicies is keyed by the basename of the container command.
var runtimePolicies = map[string]runtimePolicy{
	"podman":    {pullArgs: []string{pullNeverFlag}, check: checkByList},
	"docker":    {pullArgs: []string{pullNeverFlag}, check: checkByList},
	"container": {pullArgs: nil, check: checkByInspect},
}

// outputFunc runs a command and returns its stdout; stderr goes to the user.
type outputFunc func(ctx context.Context, name string, args ...string) ([]byte, error)

func defaultOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec
	cmd.Stderr = os.Stderr

	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running %s: %w", name, err)
	}

	return stdout, nil
}

// CheckImage returns nil when ImageName exists locally for containerCmd,
// ErrImageNotFound when it is missing, and another error when the check itself
// failed or the runtime is unsupported. It never writes to stdout.
func CheckImage(ctx context.Context, containerCmd string) error {
	return checkImageWith(ctx, defaultOutput, containerCmd)
}

func checkImageWith(ctx context.Context, output outputFunc, containerCmd string) error {
	runtime := runtimeName(containerCmd)

	policy, ok := runtimePolicies[runtime]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnsupportedRuntime, runtime)
	}

	switch policy.check {
	case checkByInspect:
		if _, err := output(ctx, containerCmd, "image", "inspect", ImageName); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("%w: %s (%s); run `chelly build` first", ErrImageNotFound, ImageName, runtime)
			}

			return fmt.Errorf("checking image with %s: %w", runtime, err)
		}

		return nil
	case checkByList:
		fallthrough
	default:
		stdout, err := output(ctx, containerCmd, "image", "ls", "-q", ImageName)
		if err != nil {
			return fmt.Errorf("checking image with %s: %w", runtime, err)
		}

		if len(bytes.TrimSpace(stdout)) == 0 {
			return fmt.Errorf("%w: %s (%s); run `chelly build` first", ErrImageNotFound, ImageName, runtime)
		}

		return nil
	}
}

// runtimeName returns the runtime identity used to look up runtime-specific behavior.
func runtimeName(containerCmd string) string {
	return filepath.Base(containerCmd)
}

// pullPolicyArgs returns the run options that stop the runtime from pulling the image.
func pullPolicyArgs(containerCmd string) []string {
	return runtimePolicies[runtimeName(containerCmd)].pullArgs
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
