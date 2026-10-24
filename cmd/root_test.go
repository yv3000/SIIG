package cmd

import (
	"encoding/json"
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

func TestRunSiig_jsonDryRunOffline(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	resetFlags(t)
	dryRun = true
	offline = true
	jsonOut = true

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := runSiig(rootCmd, nil)
	w.Close()
	os.Stdout = old
	if runErr != nil {
		t.Fatal(runErr)
	}
	buf := make([]byte, 1<<16)
	n, _ := r.Read(buf)
	raw := string(buf[:n])
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
	if strings.Contains(raw, "\x1b[") {
		t.Fatal("JSON path must not include ANSI escapes")
	}
}
