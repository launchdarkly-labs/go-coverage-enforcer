package main

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeCoverage(t *testing.T) {
	t.Run("reads blocks", func(t *testing.T) {
		expectedResult := makeTestAnalyzerExpectedResult()

		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			result, err := AnalyzeCoverage(cp, testBaseOptions)
			require.NoError(t, err)
			assert.Equal(t, expectedResult, result)
		})
	})

	t.Run("skips files based on file path pattern", func(t *testing.T) {
		expectedResult := makeTestAnalyzerExpectedResult()
		expectedResult.SkippedFilePaths = []string{testDataPackagePath + "/second"}
		expectedResult.SkippedBlocks = []CodeBlockCoverage{
			{
				CodeRange: CodeRange{
					FilePath:  testDataPackagePath + "/second",
					StartLine: 1, StartColumn: 1, EndLine: 5, EndColumn: 1,
				},
				StatementCount: 5, CoverageCount: 0,
			},
		}
		expectedResult.Packages[0].Files = []AnalyzerFileResult{
			expectedResult.Packages[0].Files[0],
			expectedResult.Packages[0].Files[2],
		}

		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			opts := testBaseOptions
			opts.SkipFilesPattern = regexp.MustCompile("econ")
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)
			assert.Equal(t, expectedResult, result)
		})
	})

	t.Run("reads code", func(t *testing.T) {
		expectedResult := makeTestAnalyzerExpectedResultWithCode()

		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			opts := testBaseOptions
			opts.ShowCode = true
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)
			assert.Equal(t, expectedResult, result)
		})
	})

	t.Run("skips blocks based on code pattern", func(t *testing.T) {
		expectedResult := makeTestAnalyzerExpectedResult()
		expectedResult.SkippedBlocks = []CodeBlockCoverage{
			{
				CodeRange: CodeRange{
					FilePath:  testDataPackagePath + "/third",
					StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1,
				},
				StatementCount: 2, CoverageCount: 0,
			},
		}
		expectedResult.Packages[0].Files[2].UncoveredBlocks = []UncoveredBlock{
			expectedResult.Packages[0].Files[2].UncoveredBlocks[1],
		}
		expectedResult.Packages[0].Files[2].TotalStatements -= 2

		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			opts := testBaseOptions
			opts.SkipCodePattern = regexp.MustCompile("third.*1")
			result, err := AnalyzeCoverage(cp, opts)
			require.NoError(t, err)
			assert.Equal(t, expectedResult, result)
		})
	})

	t.Run("error for wrong package path", func(t *testing.T) {
		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			opts := EnforcerOptions{PackagePath: "not-" + testDataPackagePath}
			_, err := AnalyzeCoverage(cp, opts)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not in the package")
		})
	})

	t.Run("error for missing source file", func(t *testing.T) {
		withValidTestProfile("coverage_data_with_bad_filename", func(cp *CoverageProfile) {
			opts := testBaseOptions
			opts.ShowCode = true
			_, err := AnalyzeCoverage(cp, opts)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unable to read file")
		})
	})
}

func TestAnalyzerWriteFilteredProfile(t *testing.T) {
	withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
		opts := testBaseOptions
		opts.SkipFilesPattern = regexp.MustCompile("econ")
		opts.SkipCodePattern = regexp.MustCompile("third.*1")
		result, err := AnalyzeCoverage(cp, opts)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		err = result.WriteFilteredProfile(cp, buf)
		require.NoError(t, err)

		cp1, err := ReadCoverageProfile(buf)
		require.NoError(t, err)

		assert.Equal(t, cp.CoverageMode, cp1.CoverageMode)
		assert.Equal(t, []CodeBlockCoverage{
			expectedParsedCoverageProfile.Blocks[0],
			expectedParsedCoverageProfile.Blocks[1],
			expectedParsedCoverageProfile.Blocks[2],
			expectedParsedCoverageProfile.Blocks[6],
			expectedParsedCoverageProfile.Blocks[7],
		}, cp1.Blocks)
	})
}
