package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yv3000/siig/cache"
	"github.com/yv3000/siig/internal/log"
	"github.com/yv3000/siig/internal/version"
	"github.com/yv3000/siig/matcher"
	"github.com/yv3000/siig/output"
	"github.com/yv3000/siig/rules"
	"github.com/yv3000/siig/scanner"
	"github.com/yv3000/siig/writer"
)

var (
	dryRun     bool
	check      bool
	jsonOut    bool
	targets    string
	verbose    bool
	offline    bool
	noCache    bool
	clearCache bool
)

// rootCmd is the base command — just run "siig" and it does the thing
var rootCmd = &cobra.Command{
	Use:   "siig",
	Short: "Auto-populate ignore files based on the detected stack",
	Long: `SIIG scans your project directory, detects the tech stack,
and automatically populates .gitignore, .vercelignore, .dockerignore,
and .vscode/settings.json with the correct entries.

Zero config. One command. No more leaked secrets.`,
	Version:      version.Version,
	SilenceUsage: true,
	RunE:         runSiig,
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be added without writing")
	rootCmd.Flags().BoolVar(&check, "check", false, "Fail (exit 1) if ignore files are missing entries, without writing")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "Write one JSON object to stdout instead of the human summary")
	rootCmd.Flags().StringVar(&targets, "targets", "", "Comma-separated targets to populate (git,vercel,docker,npm,vscode)")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Print every file checked and rule applied")
	rootCmd.Flags().BoolVar(&offline, "offline", false, "Skip all network requests, use embedded rules and cache only")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "Ignore rules cached from earlier runs and fetch fresh ones")
	rootCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "Delete rules cached from earlier runs and exit")
}

// Execute runs the root command — called from main.go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runSiig loads the embedded rules, scans the working directory, merges the
// matching rules into per-target entries, writes the applicable files and then
// reports what happened.
func runSiig(cmd *cobra.Command, args []string) error {
	// --verbose asks for more detail, so diagnostics go out as records a
	// machine can read; stdout keeps the human summary either way.
	log.SetJSON(verbose)

	if clearCache {
		if err := cache.Clear(); err != nil {
			return fmt.Errorf("could not clear cached rules: %w", err)
		}
		fmt.Println("✓ Cached rules deleted")
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	allRules, err := rules.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	scanResult, err := scanner.Scan(cwd, allRules)
	if err != nil {
		return fmt.Errorf("failed to scan project: %w", err)
	}

	// print paths the scanner could not read
	if !jsonOut {
		output.PrintScanErrors(scanResult)
	}

	var allowedTargets []matcher.Target
	if targets != "" {
		allowedTargets, err = matcher.ParseTargets(strings.Split(targets, ","))
		if err != nil {
			return err
		}
		if len(allowedTargets) == 0 {
			return fmt.Errorf("no valid targets in %q", targets)
		}
	}

	// 3. match detected stacks against rules (with online fetch if not offline)
	matchResult := matcher.Match(scanResult, allRules, allowedTargets, matcher.Options{
		ProjectRoot: cwd,
		Offline:     offline,
		UseCache:    !noCache,
	})

	// report what was applied — including rules fetched from siig-rules, which
	// the scan alone cannot know about
	if !jsonOut {
		output.PrintDetected(matchResult)
		if len(matchResult.DetectedNames)+len(matchResult.OnlineNames) == 0 {
			// still apply global rules
			fmt.Println("Applying global rules only...")
		}

		// verbose output
		if verbose {
			output.PrintVerbose(scanResult, matchResult)
		}
	}

	// 4. write to target files (--check never writes)
	noWrite := dryRun || check
	writeResults := writer.WriteAll(cwd, matchResult, noWrite)

	// 5. print results
	if jsonOut {
		if err := printJSONReport(scanResult, matchResult, writeResults); err != nil {
			return err
		}
		if check {
			if missing := pendingEntries(writeResults); missing > 0 {
				return fmt.Errorf("ignore files are out of date: %d entries missing (run `siig` to add them)", missing)
			}
		}
		return nil
	}
	switch {
	case check:
		output.PrintCheck(writeResults)
		if missing := pendingEntries(writeResults); missing > 0 {
			return fmt.Errorf("ignore files are out of date: %d entries missing (run `siig` to add them)", missing)
		}
		if hasApplicableTargets(matchResult) {
			output.PrintNoChanges()
		}
	case noWrite:
		output.PrintDryRun(writeResults)
	case pendingEntries(writeResults) == 0:
		if hasApplicableTargets(matchResult) {
			output.PrintNoChanges()
		} else {
			output.PrintResults(writeResults)
		}
	default:
		output.PrintResults(writeResults)
	}
	return nil
}

// pendingEntries counts entries that still have to be written.
func pendingEntries(results []writer.WriteResult) int {
	missing := 0
	for _, r := range results {
		if !r.Skipped {
			missing += r.EntriesAdded
		}
	}
	return missing
}

func hasApplicableTargets(result *matcher.MatchResult) bool {
	for _, applicable := range result.ApplicableTargets {
		if applicable {
			return true
		}
	}
	return false
}

// jsonTarget is one row of the machine-readable report.
type jsonTarget struct {
	Target       string `json:"target"`
	File         string `json:"file"`
	EntriesAdded int    `json:"entriesAdded"`
	Skipped      bool   `json:"skipped"`
	Reason       string `json:"reason"`
}

type jsonReport struct {
	Version      string       `json:"version"`
	Stacks       []string     `json:"stacks"`
	OnlineStacks []string     `json:"onlineStacks"`
	Targets      []jsonTarget `json:"targets"`
	ScanErrors   []string     `json:"scanErrors"`
}

func printJSONReport(scanResult *scanner.ScanResult, matchResult *matcher.MatchResult, writeResults []writer.WriteResult) error {
	report := jsonReport{
		Version:      version.Version,
		Stacks:       matchResult.DetectedNames,
		OnlineStacks: matchResult.OnlineNames,
		ScanErrors:   scanResult.ScanErrors,
	}
	if report.Stacks == nil {
		report.Stacks = []string{}
	}
	if report.OnlineStacks == nil {
		report.OnlineStacks = []string{}
	}
	if report.ScanErrors == nil {
		report.ScanErrors = []string{}
	}
	report.Targets = make([]jsonTarget, 0, len(writeResults))
	for _, r := range writeResults {
		report.Targets = append(report.Targets, jsonTarget{
			Target:       string(r.Target),
			File:         r.FilePath,
			EntriesAdded: r.EntriesAdded,
			Skipped:      r.Skipped,
			Reason:       r.SkipReason,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(report)
}
