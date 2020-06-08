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

	pass := false
	if len(result.UncoveredBlocks) == 0 {
		fmt.Println("Coverage scan passes!")
		pass = true
	} else {
		displayReport(result)
	}

	if options.OutputFilePath != "" {
		f1, err := os.Create(options.OutputFilePath)
		exitIfError(err)
		defer f1.Close()
		exitIfError(result.WriteFilteredProfile(profile, f1))
		fmt.Println("Filtered profile written to", options.OutputFilePath)
	}

	if !pass {
		os.Exit(1)
	}
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func displayReport(result AnalyzerResult) {
	fmt.Println("Uncovered blocks detected:")
	for _, b := range result.UncoveredBlocks {
		if len(b.Text) > 0 {
			fmt.Println()
		}
		fmt.Printf("%s %d-%d\n", b.CodeRange.FilePath, b.CodeRange.StartLine, b.CodeRange.EndLine)
		if len(b.Text) > 0 {
			for i, line := range b.Text {
				fmt.Printf("%d>\t%s\n", b.CodeRange.StartLine+i, line)
			}
		}
	}
}
