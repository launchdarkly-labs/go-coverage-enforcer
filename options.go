package main

import (
	"flag"
	"fmt"
	"io"
	"regexp"
)

const usageMessage = "go-coverage-enforcer [options] <coverage file>"

// EnforcerOptions is a representation of the command-line options passed to the program.
type EnforcerOptions struct {
	InputFilePath    string
	PackagePath      string
	SkipFilesPattern *regexp.Regexp
	SkipCodePattern  *regexp.Regexp
	ShowCode         bool
	OutputFilePath   string
}

// ReadCommandLineOptions parses the options from the command line. If they were invalid, it
// prints a usage message and returns false.
func ReadCommandLineOptions(argsIn []string, errWriter io.Writer) (EnforcerOptions, bool) {
	var opts EnforcerOptions

	var skipFilesPattern string
	var skipCodePattern string

	flags := flag.NewFlagSet(usageMessage, flag.ContinueOnError)
	flags.SetOutput(errWriter)
	flags.StringVar(&opts.PackagePath, "package", "", "base import path of this package")
	flags.BoolVar(&opts.ShowCode, "showcode", false, "display source code of uncovered blocks")
	flags.StringVar(&skipFilesPattern, "skipfiles", "", "regex pattern for file paths to be ignored")
	flags.StringVar(&skipCodePattern, "skipcode", "", "regex pattern for ignoring a code block")
	flags.StringVar(&opts.OutputFilePath, "outprofile", "", "save the filtered coverage profile to this path")
	err := flags.Parse(argsIn[1:])

	if err != nil {
		return opts, false
	}

	leftoverArgs := flags.Args()
	if len(leftoverArgs) != 1 {
		fmt.Fprintln(errWriter, usageMessage)
		flags.PrintDefaults()
		return opts, false
	}
	opts.InputFilePath = leftoverArgs[0]

	var ok bool
	if opts.SkipFilesPattern, ok = maybeRegexpParam(skipFilesPattern, errWriter); !ok {
		return opts, false
	}
	if opts.SkipCodePattern, ok = maybeRegexpParam(skipCodePattern, errWriter); !ok {
		return opts, false
	}

	return opts, true
}

func maybeRegexpParam(s string, errWriter io.Writer) (*regexp.Regexp, bool) {
	if s == "" {
		return nil, true
	}
	r, err := regexp.Compile(s)
	if err != nil {
		fmt.Fprintf(errWriter, "Not a valid regular expression: %s (%s)\n", s, err)
		return nil, false
	}
	return r, true
}
