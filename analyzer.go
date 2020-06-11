package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// AnalyzeCoverage applies the configured options to the profile data to produce report data.
func AnalyzeCoverage(profile *CoverageProfile, opts EnforcerOptions) (AnalyzerResult, error) {
	blocks := profile.GetUniqueBlocks()

	var result AnalyzerResult
	var currentPackage *AnalyzerPackageResult
	var currentFile *AnalyzerFileResult

	for _, b := range blocks {
		packagePath, fileName := b.CodeRange.GetPackagePathAndFileName()

		if !strings.HasPrefix(packagePath, opts.PackagePath) {
			return result, fmt.Errorf(`coverage profile refers to source files that are not in the package "%s"; use -package option to specify correct package path`,
				opts.PackagePath)
		}

		var relativePackagePath, filePath string
		if packagePath == opts.PackagePath {
			filePath = fileName
		} else {
			relativePackagePath = strings.TrimPrefix(packagePath, opts.PackagePath+"/")
			filePath = relativePackagePath + "/" + fileName
		}

		if opts.SkipFilesPattern != nil && opts.SkipFilesPattern.MatchString(filePath) {
			result.SkippedBlocks = append(result.SkippedBlocks, b)
			result.SkippedFilePaths = append(result.SkippedFilePaths, b.CodeRange.FilePath)
			continue
		}

		if currentFile == nil || fileName != currentFile.FileName {
			if currentFile != nil {
				currentPackage.Files = append(currentPackage.Files, *currentFile)
			}
			currentFile = &AnalyzerFileResult{FileName: fileName}
		}
		if currentPackage == nil || relativePackagePath != currentPackage.RelativePath {
			if currentPackage != nil {
				result.Packages = append(result.Packages, *currentPackage)
			}
			currentPackage = &AnalyzerPackageResult{RelativePath: relativePackagePath}
		}

		if b.CoverageCount > 0 {
			currentFile.TotalStatements += b.StatementCount
			currentFile.CoveredStatements += b.StatementCount
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

		currentFile.TotalStatements += b.StatementCount

		ub := UncoveredBlock{CodeRange: b.CodeRange, Text: showLines}
		currentFile.UncoveredBlocks = append(currentFile.UncoveredBlocks, ub)
	}

	if currentFile != nil {
		currentPackage.Files = append(currentPackage.Files, *currentFile)
	}
	if currentPackage != nil {
		result.Packages = append(result.Packages, *currentPackage)
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
