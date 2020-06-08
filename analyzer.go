package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// AnalyzerResult is the result returned by AnalyzeCoverage.
type AnalyzerResult struct {
	// UncoveredBlocks is a list of code ranges that had no coverage. The list is sorted in
	// ascending order of file path and line number. It does not include any locations that are
	// in SkippedBlocks or SkippedFilePaths.
	UncoveredBlocks []UncoveredBlock

	// SkippedFilePaths is a list of file paths that were skipped due to the "-skipfiles" option.
	// These are in the same format as in the coverage profile, so they include the package's
	// import path.
	SkippedFilePaths []string

	// SkippedBlocks is a list of code ranges that were skipped due to the "-skipcode" option.
	SkippedBlocks []CodeBlockCoverage
}

// UncoveredBlock is a code range that had no coverage.
type UncoveredBlock struct {
	CodeRange

	// Text is an optional excerpt of the source code file corresponding to the range's starting
	// and ending line numbers. It is only provided if the "-showcode" option was used; otherwise
	// it is nil.
	Text []string
}

// AnalyzeCoverage applies the configured options to the profile data to produce report data.
func AnalyzeCoverage(profile *CoverageProfile, opts EnforcerOptions) (AnalyzerResult, error) {
	basePathPrefix := opts.PackagePath + "/"

	blocks := profile.GetUniqueBlocks()
	var result AnalyzerResult

	for _, b := range blocks {
		if !strings.HasPrefix(b.CodeRange.FilePath, basePathPrefix) {
			return result, fmt.Errorf(`coverage profile refers to source files that are not in the package "%s"; use -package option to specify correct package path`,
				opts.PackagePath)
		}
		filePath := strings.TrimPrefix(b.CodeRange.FilePath, basePathPrefix)
		if opts.SkipFilesPattern != nil && opts.SkipFilesPattern.MatchString(filePath) {
			result.SkippedBlocks = append(result.SkippedBlocks, b)
			result.SkippedFilePaths = append(result.SkippedFilePaths, b.CodeRange.FilePath)
			continue
		}

		if b.CoverageCount > 0 {
			continue
		}

		var showLines []string
		if opts.ShowCode || opts.SkipCodePattern != nil {
			lines, err := readFileLines(filePath, b.CodeRange.StartLine, b.CodeRange.EndLine)
			if err != nil {
				return result, fmt.Errorf(`unable to read file "%s" (%s)`, filePath, err)
			}

			if opts.SkipCodePattern != nil {
				found := false
				for _, line := range lines {
					if opts.SkipCodePattern.FindString(line) != "" {
						found = true
						break
					}
				}
				if found {
					result.SkippedBlocks = append(result.SkippedBlocks, b)
					continue
				}
			}

			if opts.ShowCode {
				showLines = lines
			}
		}

		result.UncoveredBlocks = append(result.UncoveredBlocks, UncoveredBlock{
			CodeRange: b.CodeRange,
			Text:      showLines,
		})
	}

	return result, nil
}

func (ar AnalyzerResult) WriteFilteredProfile(
	originalProfile *CoverageProfile,
	writer io.Writer,
) error {
	skippedBlocksMap := make(map[CodeRange]bool, len(ar.SkippedBlocks))
	for _, s := range ar.SkippedBlocks {
		skippedBlocksMap[s.CodeRange] = true
	}
	skippedFilesMap := make(map[string]bool, len(ar.SkippedFilePaths))
	for _, s := range ar.SkippedFilePaths {
		skippedFilesMap[s] = true
	}
	filteredProfile := originalProfile.WithBlockFilter(func(b CodeBlockCoverage) bool {
		return !skippedFilesMap[b.CodeRange.FilePath] && !skippedBlocksMap[b.CodeRange]
	})
	return filteredProfile.WriteTo(writer)
}

func readFileLines(path string, start, end int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ret []string
	n := 0
	scanner := bufio.NewScanner(f)
	for n < end && scanner.Scan() {
		n++
		if n < start {
			continue
		}
		ret = append(ret, scanner.Text())
	}
	return ret, scanner.Err()
}
