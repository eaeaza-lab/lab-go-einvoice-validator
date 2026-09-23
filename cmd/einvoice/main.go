// Command einvoice validates synthetic e-invoice documents locally.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"example.com/einvoice-validator/internal/invoice"
)

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
	root.AddCommand(newVersionCmd(), newValidateCmd())
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
	var format string
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
			for _, f := range files {
				inv, err := invoice.Load(f)
				if err != nil {
					return err
				}
				diags := invoice.Validate(inv)
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
			if found {
				return errDiagnostics
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	return cmd
}
