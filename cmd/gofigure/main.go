// Command gofigure converts gofigure configuration files to JSON.
//
// Usage:
//
//	gofigure -i <input.fig> [-o <output.json>] [-indent]
//
// Flags:
//
//	-i        Input configuration file (required)
//	-o        Output JSON file (default: stdout)
//	-indent   Pretty-print JSON output with indentation
//	-version  Print version information
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shellkjell/gofigure/pkg/fig"
)

// Version information (set via ldflags).
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

const (
	exitSuccess = 0
	exitError   = 1
	exitUsage   = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gofigure", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		inputFile   string
		outputFile  string
		indent      bool
		showVersion bool
	)

	fs.StringVar(&inputFile, "i", "", "Input configuration file (required)")
	fs.StringVar(&outputFile, "o", "", "Output JSON file (default: stdout)")
	fs.BoolVar(&indent, "indent", false, "Pretty-print JSON output with indentation")
	fs.BoolVar(&showVersion, "version", false, "Print version information")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gofigure -i <input.fig> [-o <output.json>] [-indent]\n\n")
		fmt.Fprintf(stderr, "Convert gofigure configuration files to JSON.\n\n")
		fmt.Fprintf(stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(stderr, "\nExamples:\n")
		fmt.Fprintf(stderr, "  gofigure -i config.fig                 # Output to stdout\n")
		fmt.Fprintf(stderr, "  gofigure -i config.fig -o config.json  # Output to file\n")
		fmt.Fprintf(stderr, "  gofigure -i config.fig -indent         # Pretty-print output\n")
	}

	if err := fs.Parse(args); err != nil {
		return exitUsage
	}

	if showVersion {
		fmt.Fprintf(stdout, "gofigure %s (commit: %s, built: %s)\n", version, commit, buildTime)
		return exitSuccess
	}

	if inputFile == "" {
		fmt.Fprintln(stderr, "Error: input file is required")
		fs.Usage()
		return exitUsage
	}

	// Get absolute path and change to the file's directory for include resolution
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "Error: resolving path: %v\n", err)
		return exitError
	}

	dir := filepath.Dir(absPath)
	filename := filepath.Base(absPath)

	// Create parser and set working directory
	parser, err := fig.NewParser()
	if err != nil {
		fmt.Fprintf(stderr, "Error: creating parser: %v\n", err)
		return exitError
	}
	parser.SetWorkingDirectory(dir)

	// Parse the configuration
	config, err := parser.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(stderr, "Error: parsing file: %v\n", err)
		return exitError
	}

	// Transform to map
	transformer := fig.NewTransformer(parser)
	mapped, err := transformer.Transform(config)
	if err != nil {
		fmt.Fprintf(stderr, "Error: transforming config: %v\n", err)
		return exitError
	}

	// Marshal to JSON
	var output []byte
	if indent {
		output, err = json.MarshalIndent(mapped, "", "  ")
	} else {
		output, err = json.Marshal(mapped)
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error: marshaling JSON: %v\n", err)
		return exitError
	}

	// Write output
	if outputFile == "" {
		fmt.Fprintln(stdout, string(output))
	} else {
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			fmt.Fprintf(stderr, "Error: writing output file: %v\n", err)
			return exitError
		}
	}

	return exitSuccess
}
