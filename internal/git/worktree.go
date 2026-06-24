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

// Package git resolves Git repository metadata needed by chelly.
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const revParseOutputLines = 2

type revParseRunner func(dir string) (string, error)

// LinkedWorktreeCommonParent returns the parent directory of the Git common dir
// when currentDir is inside a linked worktree.
func LinkedWorktreeCommonParent(currentDir string) (string, bool) {
	return linkedWorktreeCommonParent(currentDir, runRevParse, os.Stat)
}

func linkedWorktreeCommonParent(
	currentDir string,
	runner revParseRunner,
	stat func(string) (os.FileInfo, error),
) (string, bool) {
	output, err := runner(currentDir)
	if err != nil {
		return "", false
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < revParseOutputLines {
		return "", false
	}

	gitDir := filepath.Clean(lines[0])

	commonDir := filepath.Clean(lines[1])
	if gitDir == commonDir {
		return "", false
	}

	commonParent := filepath.Dir(commonDir)

	info, err := stat(commonParent)
	if err != nil || !info.IsDir() {
		return "", false
	}

	return commonParent, true
}

func runRevParse(dir string) (string, error) {
	cmd := exec.CommandContext(
		context.Background(),
		"git",
		"rev-parse",
		"--path-format=absolute",
		"--git-dir",
		"--git-common-dir",
	)
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running git rev-parse: %w", err)
	}

	return string(output), nil
}
