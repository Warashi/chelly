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

//nolint:testpackage // Tests resolver injection seams without running git.
package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

var errGitFailed = errors.New("git failed")

func TestLinkedWorktreeCommonParent_LinkedWorktree(t *testing.T) {
	t.Parallel()

	currentDir := "/worktree"
	commonParent := t.TempDir()

	commonDir := filepath.Join(commonParent, ".git")
	if err := os.Mkdir(commonDir, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	gotDir := ""

	got, ok := linkedWorktreeCommonParent(currentDir, func(dir string) (string, error) {
		gotDir = dir

		return filepath.Join(commonDir, "worktrees", "feature") + "\n" + commonDir + "\n", nil
	}, os.Stat)
	if !ok {
		t.Fatal("linkedWorktreeCommonParent ok = false, want true")
	}

	if got != commonParent {
		t.Errorf("linkedWorktreeCommonParent path = %q, want %q", got, commonParent)
	}

	if gotDir != currentDir {
		t.Errorf("runner dir = %q, want %q", gotDir, currentDir)
	}
}

func TestLinkedWorktreeCommonParent_NormalRepository(t *testing.T) {
	t.Parallel()

	commonParent := t.TempDir()

	commonDir := filepath.Join(commonParent, ".git")
	if err := os.Mkdir(commonDir, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	got, ok := linkedWorktreeCommonParent("/repo", func(string) (string, error) {
		return commonDir + "\n" + commonDir + "\n", nil
	}, os.Stat)
	if ok {
		t.Fatalf("linkedWorktreeCommonParent ok = true, want false with path %q", got)
	}
}

func TestLinkedWorktreeCommonParent_GitFailure(t *testing.T) {
	t.Parallel()

	got, ok := linkedWorktreeCommonParent("/repo", func(string) (string, error) {
		return "", errGitFailed
	}, os.Stat)
	if ok {
		t.Fatalf("linkedWorktreeCommonParent ok = true, want false with path %q", got)
	}
}

func TestLinkedWorktreeCommonParent_MalformedOutput(t *testing.T) {
	t.Parallel()

	got, ok := linkedWorktreeCommonParent("/repo", func(string) (string, error) {
		return "/repo/.git\n", nil
	}, os.Stat)
	if ok {
		t.Fatalf("linkedWorktreeCommonParent ok = true, want false with path %q", got)
	}
}

func TestLinkedWorktreeCommonParent_InvalidCommonParent(t *testing.T) {
	t.Parallel()

	missingCommonDir := filepath.Join(t.TempDir(), "missing", ".git")

	got, ok := linkedWorktreeCommonParent("/repo", func(string) (string, error) {
		return filepath.Join(missingCommonDir, "worktrees", "feature") + "\n" + missingCommonDir + "\n", nil
	}, os.Stat)
	if ok {
		t.Fatalf("linkedWorktreeCommonParent ok = true, want false with path %q", got)
	}
}
