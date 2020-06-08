package main

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAnalyzerProfileFileName    = "test-analyzer-profile"
	testAnalyzerProfilePackagePath = "package-path"
)

var testAnalyzerProfileUncoveredBlocks = []UncoveredBlock{
	{CodeRange: CodeRange{
		FilePath:  testAnalyzerProfilePackagePath + "/file1",
		StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1,
	}},
	{CodeRange: CodeRange{
		FilePath:  testAnalyzerProfilePackagePath + "/file1",
		StartLine: 3, StartColumn: 1, EndLine: 4, EndColumn: 1,
	}},
	{CodeRange: CodeRange{
		FilePath:  testAnalyzerProfilePackagePath + "/file2-skip",
		StartLine: 1, StartColumn: 1, EndLine: 5, EndColumn: 1,
	}},
	{CodeRange: CodeRange{
		FilePath:  testAnalyzerProfilePackagePath + "/file3",
		StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1,
	}},
	{CodeRange: CodeRange{
		FilePath:  testAnalyzerProfilePackagePath + "/file3",
		StartLine: 4, StartColumn: 1, EndLine: 5, EndColumn: 1,
	}},
}

var testAnalyzerProfileUncoveredBlocksWithCode = []UncoveredBlock{
	{CodeRange: testAnalyzerProfileUncoveredBlocks[0].CodeRange,
		Text: []string{"file 1 line 1", "file 1 line 2"}},
	{CodeRange: testAnalyzerProfileUncoveredBlocks[1].CodeRange,
		Text: []string{"file 1 line 3 SKIPME", "file 1 line 4"}},
	{CodeRange: testAnalyzerProfileUncoveredBlocks[2].CodeRange,
		Text: []string{"file 2 line 1", "file 2 line 2", "file 2 line 3", "file 2 line 4", "file 2 line 5"}},
	{CodeRange: testAnalyzerProfileUncoveredBlocks[3].CodeRange,
		Text: []string{"file 3 line 1", "file 3 line 2"}},
	{CodeRange: testAnalyzerProfileUncoveredBlocks[4].CodeRange,
		Text: []string{"file 3 line 4", "file 3 line 5"}},
}

func TestAnalyzeProfile(t *testing.T) {
	t.Run("reads blocks", func(t *testing.T) {
		withValidTestProfile(testAnalyzerProfileFileName, func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath: testAnalyzerProfilePackagePath,
			}
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)

			expectedBlocks := testAnalyzerProfileUncoveredBlocks
			assert.Len(t, result.UncoveredBlocks, len(expectedBlocks))
			assert.Equal(t, expectedBlocks, result.UncoveredBlocks)

			assert.Len(t, result.SkippedFilePaths, 0)
			assert.Len(t, result.SkippedBlocks, 0)
		})
	})

	t.Run("skips files based on file path pattern", func(t *testing.T) {
		withValidTestProfile(testAnalyzerProfileFileName, func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath:      testAnalyzerProfilePackagePath,
				SkipFilesPattern: regexp.MustCompile("-skip"),
			}
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)

			expectedBlocks := []UncoveredBlock{
				testAnalyzerProfileUncoveredBlocks[0],
				testAnalyzerProfileUncoveredBlocks[1],
				testAnalyzerProfileUncoveredBlocks[3],
				testAnalyzerProfileUncoveredBlocks[4],
			}
			assert.Len(t, result.UncoveredBlocks, len(expectedBlocks))
			assert.Equal(t, expectedBlocks, result.UncoveredBlocks)

			assert.Len(t, result.SkippedFilePaths, 1)
			assert.Equal(t, []string{testAnalyzerProfilePackagePath + "/file2-skip"}, result.SkippedFilePaths)
			assert.Len(t, result.SkippedBlocks, 1)
			assert.Equal(t, testAnalyzerProfileUncoveredBlocks[2].CodeRange, result.SkippedBlocks[0].CodeRange)
		})
	})

	t.Run("reads code", func(t *testing.T) {
		withValidTestProfile(testAnalyzerProfileFileName, func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath: testAnalyzerProfilePackagePath,
				ShowCode:    true,
			}
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)

			expectedBlocks := testAnalyzerProfileUncoveredBlocksWithCode
			assert.Len(t, result.UncoveredBlocks, len(expectedBlocks))
			assert.Equal(t, expectedBlocks, result.UncoveredBlocks)
		})
	})

	t.Run("skips blocks based on code pattern", func(t *testing.T) {
		withValidTestProfile(testAnalyzerProfileFileName, func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath:     testAnalyzerProfilePackagePath,
				SkipCodePattern: regexp.MustCompile("SKIPME"),
			}
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)

			buf := new(bytes.Buffer)
			err = result.WriteFilteredProfile(cp, buf)
			require.NoError(t, err)

			expectedBlocks := []UncoveredBlock{
				testAnalyzerProfileUncoveredBlocks[0],
				testAnalyzerProfileUncoveredBlocks[2],
				testAnalyzerProfileUncoveredBlocks[3],
				testAnalyzerProfileUncoveredBlocks[4],
			}
			assert.Len(t, result.UncoveredBlocks, len(expectedBlocks))
			assert.Equal(t, expectedBlocks, result.UncoveredBlocks)

			assert.Len(t, result.SkippedFilePaths, 0)
			assert.Len(t, result.SkippedBlocks, 1)
			assert.Equal(t, testAnalyzerProfileUncoveredBlocks[1].CodeRange, result.SkippedBlocks[0].CodeRange)
		})
	})

	t.Run("error for wrong package path", func(t *testing.T) {
		withValidTestProfile(testAnalyzerProfileFileName, func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath: "not-" + testAnalyzerProfilePackagePath,
			}
			_, err := AnalyzeCoverage(cp, opts)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not in the package")
		})
	})

	t.Run("error for missing source file", func(t *testing.T) {
		withValidTestProfile("test-analyzer-profile-with-bad-filename", func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath: testAnalyzerProfilePackagePath,
				ShowCode:    true,
			}
			_, err := AnalyzeCoverage(cp, opts)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unable to read file")
		})
	})

	t.Run("reads blocks", func(t *testing.T) {

	})
}

func TestAnalyzerResultMethods(t *testing.T) {
	t.Run("WriteFilteredProfile", func(t *testing.T) {
		withValidTestProfile(testAnalyzerProfileFileName, func(cp *CoverageProfile) {
			opts := EnforcerOptions{
				PackagePath:      testAnalyzerProfilePackagePath,
				SkipFilesPattern: regexp.MustCompile("-skip"),
				SkipCodePattern:  regexp.MustCompile("SKIPME"),
			}
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)

			buf := new(bytes.Buffer)
			err = result.WriteFilteredProfile(cp, buf)
			require.NoError(t, err)

			cp1, err := ReadCoverageProfile(buf)
			require.NoError(t, err)

			assert.Equal(t, cp.CoverageMode, cp1.CoverageMode)
			assert.Len(t, cp1.Blocks, 5)
			assert.Equal(t, testAnalyzerProfileUncoveredBlocks[0].CodeRange, cp1.Blocks[0].CodeRange)
			assert.Equal(t, testAnalyzerProfileUncoveredBlocks[3].CodeRange, cp1.Blocks[1].CodeRange)
			// The next 2 blocks are not part of testAnalyzerProfileUncoveredBlocks because they represent a
			// range that *did* have coverage (eventually); it's not WriteFilteredProfile's job to filter those out.
			assert.Equal(t,
				CodeBlockCoverage{
					CodeRange: CodeRange{
						FilePath:  testAnalyzerProfilePackagePath + "/file3",
						StartLine: 3, StartColumn: 1, EndLine: 4, EndColumn: 1,
					},
					StatementCount: 2,
					CoverageCount:  0,
				},
				cp1.Blocks[2],
			)
			assert.Equal(t,
				CodeBlockCoverage{
					CodeRange: CodeRange{
						FilePath:  testAnalyzerProfilePackagePath + "/file3",
						StartLine: 3, StartColumn: 1, EndLine: 4, EndColumn: 1,
					},
					StatementCount: 2,
					CoverageCount:  1,
				},
				cp1.Blocks[3],
			)
			assert.Equal(t, testAnalyzerProfileUncoveredBlocks[4].CodeRange, cp1.Blocks[4].CodeRange)
		})
	})
}
