package main

import (
	"fmt"
	"os"
)

func main() {
	options, ok := ReadCommandLineOptions(os.Args, os.Stderr)
	if !ok {
		os.Exit(1)
	}

	if options.PackagePath == "" {
		options.PackagePath = InferPackagePath()
		if options.PackagePath == "" {
			fmt.Fprintln(os.Stderr, "Unable to determine package path; use -package option")
			os.Exit(1)
		}
	}

	f, err := os.Open(options.InputFilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to read", options.InputFilePath)
		os.Exit(1)
	}
	defer f.Close()
	profile, err := ReadCoverageProfile(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading profile:", err)
		os.Exit(1)
	}

	result, err := AnalyzeCoverage(profile, options)
	exitIfError(err)

	report := NewSummaryReport(result, options)
	report.Output(os.Stdout, options)

	if options.OutputFilePath != "" {
		f1, err := os.Create(options.OutputFilePath)
		exitIfError(err)
		defer f1.Close()
		exitIfError(result.WriteFilteredProfile(profile, f1))
		fmt.Println("Filtered profile written to", options.OutputFilePath)
	}

	if !report.Pass {
		os.Exit(1)
	}
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
