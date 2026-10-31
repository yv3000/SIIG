package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yv3000/siig/cache"
	"github.com/yv3000/siig/internal/version"
	"github.com/yv3000/siig/rules"
)

// resetFlags clears the package-level flag vars before a test and again afterwards,
// so cases neither inherit nor leak flags.
func resetFlags(t *testing.T) {
	t.Helper()

	clear := func() {
		dryRun, check, targets, verbose = false, false, "", false
		offline, noCache, clearCache, jsonOut = false, false, false, false
	}
	clear()
	t.Cleanup(clear)
}

// isolateHome points the cache at a throwaway directory.
func isolateHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}

func TestRootCmdUsesSharedVersion(t *testing.T) {
	if rootCmd.Version != version.Version {
		t.Fatalf("rootCmd.Version = %q, want %q", rootCmd.Version, version.Version)
	}
}

func TestRunSiig_dryRunOffline(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	resetFlags(t)
	dryRun = true
	offline = true
	if err := runSiig(rootCmd, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil {
		t.Fatal("dry-run must not write .gitignore")
	}
}

func TestRunSiig_badTargets(t *testing.T) {
	chdir(t, t.TempDir())

	resetFlags(t)
	dryRun = true
	offline = true
	targets = "nope"
	if err := runSiig(rootCmd, nil); err == nil {
		t.Fatal("expected unknown target error")
	}
}

func TestRunSiig_clearCache(t *testing.T) {
	isolateHome(t)
	if err := cache.Set("gleam", &rules.StackRule{ID: "gleam", Name: "Gleam"}); err != nil {
		t.Fatal(err)
	}
	if !cache.IsCached("gleam") {
		t.Fatal("test setup: expected a cached rule")
	}

	resetFlags(t)
	clearCache = true
	if err := runSiig(rootCmd, nil); err != nil {
		t.Fatal(err)
	}
	if cache.IsCached("gleam") {
		t.Fatal("--clear-cache should remove cached rules")
	}
}

func TestRunSiig_noCacheKeepsCacheOnDisk(t *testing.T) {
	isolateHome(t)
	chdir(t, t.TempDir())
	if err := cache.Set("gleam", &rules.StackRule{ID: "gleam", Name: "Gleam"}); err != nil {
		t.Fatal(err)
	}

	resetFlags(t)
	dryRun = true
	offline = true
	noCache = true
	if err := runSiig(rootCmd, nil); err != nil {
		t.Fatal(err)
	}
	if !cache.IsCached("gleam") {
		t.Fatal("--no-cache should ignore the cache, not delete it")
	}
}

func TestRunSiig_checkFailsUntilIgnoreFilesAreCurrent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)
	isolateHome(t)

	resetFlags(t)
	offline, check = true, true
	if err := runSiig(rootCmd, nil); err == nil {
		t.Fatal("--check should fail while entries are missing")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil {
		t.Fatal("--check must not write")
	}

	resetFlags(t)
	offline = true
	if err := runSiig(rootCmd, nil); err != nil {
		t.Fatal(err)
	}

	resetFlags(t)
	offline, check = true, true
	if err := runSiig(rootCmd, nil); err != nil {
		t.Fatalf("--check should pass once ignore files are current: %v", err)
	}
}

func TestRunSiig_nonApplicableTargetIsNotReportedUpToDate(t *testing.T) {
	for _, checkMode := range []bool{false, true} {
		name := "run"
		if checkMode {
			name = "check"
		}
		t.Run(name, func(t *testing.T) {
			chdir(t, t.TempDir())
			isolateHome(t)
			resetFlags(t)
			offline = true
			targets = "docker"
			check = checkMode

			var err error
			out := captureStdout(t, func() {
				err = runSiig(rootCmd, nil)
			})
			if err != nil {
				t.Fatalf("non-applicable target should not be an error: %v", err)
			}
			if strings.Contains(out, "Everything already up to date") {
				t.Fatalf("non-applicable target must not be reported as current:\n%s", out)
			}
			if !strings.Contains(out, ".dockerignore") || !strings.Contains(out, "not applicable for this project") {
				t.Fatalf("output should explain that Docker is not applicable:\n%s", out)
			}
		})
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	if err := w.Close(); err != nil {
		os.Stdout = original
		t.Fatal(err)
	}
	os.Stdout = original

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestRunSiig_jsonOutput(t *testing.T) {
	t.Run("dry-run emits valid JSON with bare filenames", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644); err != nil {
			t.Fatal(err)
		}
		chdir(t, dir)
		isolateHome(t)

		resetFlags(t)
		dryRun = true
		offline = true
		jsonOut = true

		var runErr error
		raw := captureStdout(t, func() {
			runErr = runSiig(rootCmd, nil)
		})
		if runErr != nil {
			t.Fatal(runErr)
		}
		if strings.Contains(raw, "\x1b[") {
			t.Fatal("JSON output must not include ANSI escapes")
		}
		var report jsonReport
		if err := json.Unmarshal([]byte(raw), &report); err != nil {
			t.Fatalf("stdout is not JSON: %v\n%s", err, raw)
		}
		if report.Version != version.Version {
			t.Fatalf("version = %q, want %q", report.Version, version.Version)
		}
		if report.Stacks == nil || report.OnlineStacks == nil || report.ScanErrors == nil || report.Targets == nil {
			t.Fatalf("nil slice in report: %+v", report)
		}
		for _, row := range report.Targets {
			if filepath.IsAbs(row.File) || strings.Contains(row.File, dir) {
				t.Fatalf("target %q leaked absolute path in File = %q", row.Target, row.File)
			}
		}
	})

	t.Run("check fails while still writing valid JSON", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644); err != nil {
			t.Fatal(err)
		}
		chdir(t, dir)
		isolateHome(t)

		resetFlags(t)
		check = true
		offline = true
		jsonOut = true

		var runErr error
		raw := captureStdout(t, func() {
			runErr = runSiig(rootCmd, nil)
		})
		if runErr == nil {
			t.Fatal("--json --check should return an error when entries are missing")
		}
		var report jsonReport
		if err := json.Unmarshal([]byte(raw), &report); err != nil {
			t.Fatalf("stdout is not JSON on --check failure: %v\n%s", err, raw)
		}
		if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil {
			t.Fatal("--json --check must not write .gitignore")
		}
	})
}
