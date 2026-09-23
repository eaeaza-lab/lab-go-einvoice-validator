// Command einvoice validates synthetic e-invoice documents locally.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"example.com/einvoice-validator/internal/diag"
	"example.com/einvoice-validator/internal/fixtures"
	"example.com/einvoice-validator/internal/invoice"
	"example.com/einvoice-validator/internal/store"
)

// defaultDB is the history file `einvoice history` reads when --db is not given.
const defaultDB = "einvoice-history.db"

// Version is the CLI version reported by `einvoice version`.
const Version = "0.1.0"

// Exit codes: 0 all valid, 1 diagnostics found, 2 usage or I/O error.
const (
	exitOK          = 0
	exitDiagnostics = 1
	exitError       = 2
)

// errDiagnostics signals that validation completed and found problems.
var errDiagnostics = errors.New("diagnostics found")

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI with the given arguments and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	err := cmd.Execute()
	switch {
	case err == nil:
		return exitOK
	case errors.Is(err, errDiagnostics):
		return exitDiagnostics
	default:
		fmt.Fprintln(stderr, "error:", err)
		return exitError
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "einvoice",
		Short:         "Validate synthetic e-invoice documents offline",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd(), newValidateCmd(), newHistoryCmd(), newDemoCmd(), newExplainCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "einvoice", Version)
			return nil
		},
	}
}

// fileDiagnostic is a diagnostic tagged with the file it came from (JSON output).
type fileDiagnostic struct {
	File string `json:"file"`
	invoice.Diagnostic
}

// jsonReport is the document printed by `validate --format json`.
type jsonReport struct {
	Valid       bool             `json:"valid"`
	Files       []string         `json:"files"`
	Diagnostics []fileDiagnostic `json:"diagnostics"`
}

func newValidateCmd() *cobra.Command {
	var format, dbPath string
	cmd := &cobra.Command{
		Use:   "validate <file>...",
		Short: "Validate one or more JSON invoice files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, files []string) error {
			if format != "text" && format != "json" {
				return fmt.Errorf("unsupported --format %q (want text or json)", format)
			}
			out := cmd.OutOrStdout()
			report := jsonReport{Files: files, Diagnostics: []fileDiagnostic{}}
			found := false
			total := 0
			for _, f := range files {
				inv, err := invoice.Load(f)
				if err != nil {
					return err
				}
				diags := invoice.Validate(inv)
				total += len(diags)
				if format == "json" {
					for _, d := range diags {
						report.Diagnostics = append(report.Diagnostics, fileDiagnostic{File: f, Diagnostic: d})
					}
					continue
				}
				if len(diags) == 0 {
					fmt.Fprintf(out, "%s: OK\n", f)
					continue
				}
				found = true
				for _, d := range diags {
					fmt.Fprintf(out, "%s: %s %s: %s\n", f, d.Code, d.Path, d.Message)
					if d.Fix != "" {
						fmt.Fprintf(out, "    fix: %s\n", d.Fix)
					}
				}
			}
			if format == "json" {
				found = len(report.Diagnostics) > 0
				report.Valid = !found
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(report); err != nil {
					return err
				}
			}
			if dbPath != "" {
				run := store.Run{Files: files, Valid: !found, Diagnostics: total}
				if err := record(dbPath, run); err != nil {
					return err
				}
			}
			if found {
				return errDiagnostics
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().StringVar(&dbPath, "db", "", "record this run in the SQLite history file at this path")
	return cmd
}

// record appends a run to the history database at path.
func record(path string, r store.Run) error {
	s, err := store.Open(path)
	if err != nil {
		return err
	}
	defer s.Close()
	_, err = s.Add(r)
	return err
}

func newDemoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "demo",
		Short: "Validate the embedded synthetic fixtures and show the diagnostics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			all, err := fixtures.All()
			if err != nil {
				return err
			}
			mismatch := false
			for _, fx := range all {
				inv, err := invoice.ParseJSON(fx.Data)
				if err != nil {
					return fmt.Errorf("fixture %s: %w", fx.Name, err)
				}
				diags := invoice.Validate(inv)
				codes := make([]string, len(diags))
				for i, d := range diags {
					codes[i] = d.Code
				}
				if strings.Join(codes, ",") != strings.Join(fx.Expected, ",") {
					mismatch = true
					fmt.Fprintf(out, "%s: UNEXPECTED (got [%s], want [%s])\n",
						fx.Name, strings.Join(codes, " "), strings.Join(fx.Expected, " "))
				}
				if len(diags) == 0 {
					fmt.Fprintf(out, "%s: OK\n", fx.Name)
					continue
				}
				for _, d := range diags {
					fmt.Fprintf(out, "%s: %s %s: %s\n", fx.Name, d.Code, d.Path, d.Message)
					if d.Fix != "" {
						fmt.Fprintf(out, "    fix: %s\n", d.Fix)
					}
				}
			}
			if mismatch {
				return errDiagnostics
			}
			return nil
		},
	}
}

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain [code]",
		Short: "Explain a diagnostic code (no argument lists all codes)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				for _, e := range diag.All() {
					fmt.Fprintf(out, "%s: %s\n", e.Code, e.Summary)
				}
				return nil
			}
			e, ok := diag.Lookup(strings.ToUpper(args[0]))
			if !ok {
				return fmt.Errorf("unknown diagnostic code %q", args[0])
			}
			fmt.Fprintf(out, "%s\n  %s\n  fix: %s\n", e.Code, e.Summary, e.Fix)
			return nil
		},
	}
}

func newHistoryCmd() *cobra.Command {
	var dbPath string
	var limit int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "List recorded validation runs, newest first",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if limit < 0 {
				return fmt.Errorf("--limit must not be negative")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return err
			}
			defer s.Close()
			runs, err := s.List(limit)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(runs) == 0 {
				fmt.Fprintln(out, "no runs recorded")
				return nil
			}
			for _, r := range runs {
				status := "OK"
				if !r.Valid {
					status = "FAIL"
				}
				fmt.Fprintf(out, "#%d %s %s diagnostics=%d files=%s\n", r.ID,
					r.Time.UTC().Format(time.RFC3339), status, r.Diagnostics, strings.Join(r.Files, ","))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db", defaultDB, "SQLite history file")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum runs to show (0 = all)")
	return cmd
}
