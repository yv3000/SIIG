package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yv3000/siig/internal/version"
	"github.com/yv3000/siig/matcher"
)

// WriteResult holds what happened when writing to one target
type WriteResult struct {
	Target       matcher.Target
	FilePath     string
	EntriesAdded int
	Skipped      bool
	SkipReason   string
	Created      bool // file was created fresh
}

// ignoreFiles lists the append-style targets in the order they are reported.
// A fixed order keeps CLI output stable between runs, which map iteration does not.
var ignoreFiles = []struct {
	target   matcher.Target
	filename string
}{
	{matcher.TargetGitignore, ".gitignore"},
	{matcher.TargetVercelignore, ".vercelignore"},
	{matcher.TargetDockerignore, ".dockerignore"},
	{matcher.TargetNpmignore, ".npmignore"},
}

// WriteAll writes entries to all applicable target files. Each file is read
// first so only entries it does not already contain get appended.
func WriteAll(projectRoot string, matchResult *matcher.MatchResult, dryRun bool) []WriteResult {
	var results []WriteResult

	for _, target := range ignoreFiles {
		filePath := filepath.Join(projectRoot, target.filename)
		if !matchResult.ApplicableTargets[target.target] {
			results = append(results, WriteResult{
				Target:     target.target,
				FilePath:   target.filename,
				Skipped:    true,
				SkipReason: "not applicable for this project",
			})
			continue
		}

		entries := matchResult.Entries[target.target]
		if len(entries) == 0 {
			results = append(results, WriteResult{
				Target:     target.target,
				FilePath:   target.filename,
				Skipped:    true,
				SkipReason: "no entries to add",
			})
			continue
		}

		results = append(results, writeIgnoreFile(filePath, entries, target.target, dryRun))
	}

	// handle vscode settings
	reason := ""
	switch {
	case !matchResult.ApplicableTargets[matcher.TargetVSCode]:
		reason = "no .vscode directory found"
	case len(matchResult.VSCodeExclude) == 0:
		reason = "no entries to add"
	}

	if reason != "" {
		results = append(results, WriteResult{
			Target:     matcher.TargetVSCode,
			FilePath:   ".vscode/settings.json",
			Skipped:    true,
			SkipReason: reason,
		})
		return results
	}

	filePath := filepath.Join(projectRoot, ".vscode", "settings.json")
	results = append(results, writeVSCodeSettings(filePath, matchResult.VSCodeExclude, dryRun))

	return results
}

// writeIgnoreFile appends new entries to an ignore-style file
func writeIgnoreFile(filePath string, entries []string, target matcher.Target, dryRun bool) WriteResult {
	result := WriteResult{
		Target:   target,
		FilePath: filepath.Base(filePath),
	}

	existing := readIgnoreFile(filePath)

	// figure out what's actually new
	var newEntries []string
	for _, entry := range entries {
		if !existing.entries[entry] {
			newEntries = append(newEntries, entry)
		}
	}

	if len(newEntries) == 0 {
		result.Skipped = true
		result.SkipReason = "all entries already present"
		return result
	}

	result.EntriesAdded = len(newEntries)
	result.Created = !existing.exists

	if dryRun {
		return result
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("failed to open file: %v", err)
		return result
	}

	if _, err := f.WriteString(ignoreBlock(newEntries, existing)); err != nil {
		_ = f.Close()
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("failed to write: %v", err)
		return result
	}

	// A failed close can lose the appended block, so it is not ignorable here.
	if err := f.Close(); err != nil {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("failed to close: %v", err)
		return result
	}

	return result
}

// ignoreBlock renders the lines to append. The SIIG header is written once per
// file — re-emitting it buried a duplicate marker under the file on every run
// that added anything.
func ignoreBlock(newEntries []string, existing ignoreFile) string {
	var builder strings.Builder

	if !existing.hasMarker {
		if existing.size > 0 {
			if !existing.endsWithNewline {
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
		builder.WriteString(version.Header())
		builder.WriteString("\n")
	} else if !existing.endsWithNewline {
		builder.WriteString("\n")
	}

	for _, entry := range newEntries {
		builder.WriteString(entry)
		builder.WriteString("\n")
	}

	return builder.String()
}

// ignoreFile is what a target file looked like before we touched it.
type ignoreFile struct {
	entries         map[string]bool
	exists          bool
	hasMarker       bool
	size            int
	endsWithNewline bool
}

// readIgnoreFile loads an ignore file, tolerating one that does not exist yet.
func readIgnoreFile(filePath string) ignoreFile {
	file := ignoreFile{entries: make(map[string]bool), endsWithNewline: true}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return file
	}

	file.exists = true
	file.size = len(data)
	file.hasMarker = strings.Contains(string(data), version.Marker)
	if file.size > 0 {
		file.endsWithNewline = data[file.size-1] == '\n'
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			file.entries[line] = true
		}
	}

	return file
}

// writeVSCodeSettings merges files.exclude into .vscode/settings.json
func writeVSCodeSettings(filePath string, excludes map[string]bool, dryRun bool) WriteResult {
	result := WriteResult{
		Target:   matcher.TargetVSCode,
		FilePath: ".vscode/settings.json",
	}

	// read existing settings
	var settings map[string]interface{}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			settings = make(map[string]interface{})
			result.Created = true
		} else {
			result.Skipped = true
			result.SkipReason = fmt.Sprintf("failed to read: %v", err)
			return result
		}
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			// file exists but isn't valid JSON — don't touch it
			result.Skipped = true
			result.SkipReason = "existing file is not valid JSON"
			return result
		}
	}

	// get or create files.exclude
	var filesExclude map[string]interface{}
	if existing, ok := settings["files.exclude"]; ok {
		if m, ok := existing.(map[string]interface{}); ok {
			filesExclude = m
		} else {
			filesExclude = make(map[string]interface{})
		}
	} else {
		filesExclude = make(map[string]interface{})
	}

	// count how many new entries we're adding
	added := 0
	for key, val := range excludes {
		if _, exists := filesExclude[key]; !exists {
			filesExclude[key] = val
			added++
		}
	}

	if added == 0 {
		result.Skipped = true
		result.SkipReason = "all entries already present"
		return result
	}

	result.EntriesAdded = added

	if dryRun {
		return result
	}

	settings["files.exclude"] = filesExclude

	// write back with nice formatting
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("failed to marshal JSON: %v", err)
		return result
	}
	output = append(output, '\n')

	// ensure .vscode directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("failed to create directory: %v", err)
		return result
	}

	if err := os.WriteFile(filePath, output, 0644); err != nil {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("failed to write: %v", err)
		return result
	}

	return result
}
