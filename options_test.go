package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func forValidCommandLine(t *testing.T, args string, action func(opts EnforcerOptions)) {
	buf := new(bytes.Buffer)
	opts, ok := ReadCommandLineOptions(strings.Split(args, " "), buf)
	if assert.True(t, ok) && assert.Equal(t, "", buf.String()) {
		action(opts)
	}
}

func forInvalidCommandLine(t *testing.T, args string, action func(errorOutput string)) {
	buf := new(bytes.Buffer)
	_, ok := ReadCommandLineOptions(strings.Split(args, " "), buf)
	if assert.False(t, ok) {
		action(buf.String())
	}
}

func TestReadCommandLineOptions(t *testing.T) {
	t.Run("valid defaults", func(t *testing.T) {
		forValidCommandLine(t, "enforcer param1", func(opts EnforcerOptions) {
			assert.Equal(t, "param1", opts.InputFilePath)
			assert.Equal(t, "", opts.PackagePath)
			assert.Nil(t, opts.SkipFilesPattern)
			assert.Nil(t, opts.SkipCodePattern)
			assert.False(t, opts.ShowCode)
			assert.Equal(t, "", opts.OutputFilePath)
		})
	})

	t.Run("-packagepath", func(t *testing.T) {
		forValidCommandLine(t, "enforcer -package example.com/my/path param1", func(opts EnforcerOptions) {
			assert.Equal(t, "example.com/my/path", opts.PackagePath)
		})
	})

	t.Run("-skipfiles", func(t *testing.T) {
		forValidCommandLine(t, "enforcer -skipfiles skip.*go param1", func(opts EnforcerOptions) {
			assert.Equal(t, regexp.MustCompile("skip.*go"), opts.SkipFilesPattern)
		})

		forInvalidCommandLine(t, "enforcer -skipfiles ??? param1", func(errorOutput string) {
			assert.Contains(t, errorOutput, "Not a valid regular expression")
		})
	})

	t.Run("-skipcode", func(t *testing.T) {
		forValidCommandLine(t, "enforcer -skipcode NOTME param1", func(opts EnforcerOptions) {
			assert.Equal(t, regexp.MustCompile("NOTME"), opts.SkipCodePattern)
		})

		forInvalidCommandLine(t, "enforcer -skipcode ??? param1", func(errorOutput string) {
			assert.Contains(t, errorOutput, "Not a valid regular expression")
		})
	})

	t.Run("-showcode", func(t *testing.T) {
		forValidCommandLine(t, "enforcer -showcode param1", func(opts EnforcerOptions) {
			assert.True(t, opts.ShowCode)
		})

		forValidCommandLine(t, "enforcer -showcode=true param1", func(opts EnforcerOptions) {
			assert.True(t, opts.ShowCode)
		})

		forValidCommandLine(t, "enforcer -showcode=false param1", func(opts EnforcerOptions) {
			assert.False(t, opts.ShowCode)
		})

		forInvalidCommandLine(t, "enforcer -showcode=3 param1", func(errorOutput string) {
			assert.Contains(t, errorOutput, "invalid boolean value")
		})
	})

	t.Run("-outprofile", func(t *testing.T) {
		forValidCommandLine(t, "enforcer -outprofile newfile param1", func(opts EnforcerOptions) {
			assert.Equal(t, "newfile", opts.OutputFilePath)
		})
	})

	t.Run("not enough params", func(t *testing.T) {
		forInvalidCommandLine(t, "enforcer", func(errorOutput string) {
			assert.Contains(t, errorOutput, "go-coverage-enforcer [options]")
		})
	})

	t.Run("too many params", func(t *testing.T) {
		forInvalidCommandLine(t, "enforcer param1 param2", func(errorOutput string) {
			assert.Contains(t, errorOutput, "go-coverage-enforcer [options]")
		})
	})

	t.Run("unknown option", func(t *testing.T) {
		forInvalidCommandLine(t, "enforcer -whatever param1", func(errorOutput string) {
			assert.Contains(t, errorOutput, "not defined: -whatever")
			assert.Contains(t, errorOutput, "go-coverage-enforcer [options]")
		})
	})
}
