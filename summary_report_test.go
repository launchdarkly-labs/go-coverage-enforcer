package main

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataReportNotPassFile = "coverage_data_for_report_not_pass"
	testDataReportPassFile    = "coverage_data_for_report_pass"
)

func TestNewSummaryReport(t *testing.T) {
	t.Run("report from data with coverage gaps", func(t *testing.T) {
		withValidTestProfile(testDataReportNotPassFile, func(cp *CoverageProfile) {
			result, err := AnalyzeCoverage(cp, testBaseOptions)
			require.NoError(t, err)
			report := NewSummaryReport(result, testBaseOptions)

			assert.False(t, report.Pass)

			assert.Len(t, report.Packages, 2)

			p1 := report.Packages[0]
			assert.Equal(t, testDataPackagePath, p1.FullPackagePath)
			assert.Equal(t, SummaryReportCoverage{8, 6}, p1.Coverage)
			assert.Equal(t, 75, p1.Coverage.GetCoveredPercent())
			assert.Len(t, p1.Files, 3)

			p1f1 := p1.Files[0]
			assert.Equal(t, "first", p1f1.FileName)
			assert.Equal(t, SummaryReportCoverage{4, 2}, p1f1.Coverage)
			assert.Equal(t, 50, p1f1.Coverage.GetCoveredPercent())

			p1f2 := p1.Files[1]
			assert.Equal(t, "second", p1f2.FileName)
			assert.Equal(t, SummaryReportCoverage{4, 4}, p1f2.Coverage)
			assert.Equal(t, 100, p1f2.Coverage.GetCoveredPercent())

			p1f3 := p1.Files[2]
			assert.Equal(t, "third", p1f3.FileName)
			assert.Equal(t, SummaryReportCoverage{0, 0}, p1f3.Coverage)
			// It shouldn't be possible for a block to have 0 statements, but this verifies that if such a
			// thing happened, we wouldn't get a divide-by-zero error.
			assert.Equal(t, 100, p1f3.Coverage.GetCoveredPercent())

			p2 := report.Packages[1]
			assert.Equal(t, testDataPackagePath+"/otherpackage", p2.FullPackagePath)
			assert.Equal(t, SummaryReportCoverage{7, 0}, p2.Coverage)
			assert.Equal(t, 0, p2.Coverage.GetCoveredPercent())
			assert.Len(t, p2.Files, 1)

			p2f1 := p2.Files[0]
			assert.Equal(t, "first", p2f1.FileName)
			assert.Equal(t, SummaryReportCoverage{7, 0}, p2f1.Coverage)
			assert.Equal(t, 0, p2f1.Coverage.GetCoveredPercent())

			var blocks []UncoveredBlock
			blocks = append(blocks, result.Packages[0].Files[0].UncoveredBlocks...)
			blocks = append(blocks, result.Packages[0].Files[1].UncoveredBlocks...)
			blocks = append(blocks, result.Packages[1].Files[0].UncoveredBlocks...)
			assert.Equal(t, blocks, report.UncoveredBlocks)
		})
	})

	t.Run("report from data with no coverage gaps", func(t *testing.T) {
		withValidTestProfile(testDataReportPassFile, func(cp *CoverageProfile) {
			result, err := AnalyzeCoverage(cp, testBaseOptions)
			require.NoError(t, err)
			report := NewSummaryReport(result, testBaseOptions)

			assert.True(t, report.Pass)

			assert.Len(t, report.Packages, 1)

			p1 := report.Packages[0]
			assert.Equal(t, testDataPackagePath, p1.FullPackagePath)
			assert.Equal(t, SummaryReportCoverage{6, 6}, p1.Coverage)
			assert.Equal(t, 100, p1.Coverage.GetCoveredPercent())
			assert.Len(t, p1.Files, 2)

			p1f1 := p1.Files[0]
			assert.Equal(t, "first", p1f1.FileName)
			assert.Equal(t, SummaryReportCoverage{2, 2}, p1f1.Coverage)
			assert.Equal(t, 100, p1f1.Coverage.GetCoveredPercent())

			p1f2 := p1.Files[1]
			assert.Equal(t, "second", p1f2.FileName)
			assert.Equal(t, SummaryReportCoverage{4, 4}, p1f2.Coverage)
			assert.Equal(t, 100, p1f2.Coverage.GetCoveredPercent())

			assert.Len(t, report.UncoveredBlocks, 0)
		})
	})
}

func TestReportOutput(t *testing.T) {
	opts := testBaseOptions
	for _, opts.ShowPackageStats = range []bool{false, true} {
		for _, opts.ShowFileStats = range []bool{false, true} {
			for _, opts.ShowCode = range []bool{false, true} {
				for _, params := range reportOutputTests {
					withValidTestProfile(params.fileName, func(cp *CoverageProfile) {
						t.Run(
							fmt.Sprintf("packagestats=%t -filestats=%t -showcode=%t",
								opts.ShowPackageStats, opts.ShowFileStats, opts.ShowCode),
							func(t *testing.T) { doReportOutputTest(t, cp, opts, params) })
					})
				}
			}
		}
	}
}

func doReportOutputTest(t *testing.T, cp *CoverageProfile, opts EnforcerOptions, testParams reportOutputTestParams) {
	result, err := AnalyzeCoverage(cp, opts)
	require.NoError(t, err)
	report := NewSummaryReport(result, opts)

	expected := ""

	if opts.ShowPackageStats {
		if opts.ShowFileStats {
			expected += testParams.textWithPackageStatsAndFileStats + "\n\n"
		} else {
			expected += testParams.textWithPackageStats + "\n\n"
		}
	} else if opts.ShowFileStats {
		expected += testParams.textWithFileStats + "\n\n"
	}

	if testParams.shouldPass {
		expected += "Coverage scan passes!\n"
	} else {
		if opts.ShowCode {
			expected += `Uncovered blocks detected:

base-package/first 3-4
3>	first file line 3
4>	first file line 4

base-package/otherpackage/first 1-3
1>	other package first file line 1
2>	other package first file line 2
3>	other package first file line 3
`
		} else {
			expected += `Uncovered blocks detected:
base-package/first 3-4
base-package/otherpackage/first 1-3
`
		}
	}

	buf := new(bytes.Buffer)
	report.Output(buf, opts)
	s := buf.String()
	s = regexp.MustCompile("  *").ReplaceAllString(s, " ")
	s = regexp.MustCompile(" \n").ReplaceAllString(s, "\n")

	assert.Equal(t, expected, s)
}

type reportOutputTestParams struct {
	shouldPass                       bool
	fileName                         string
	textWithPackageStatsAndFileStats string
	textWithPackageStats             string
	textWithFileStats                string
}

var reportOutputTests = []reportOutputTestParams{
	{
		shouldPass: false,
		fileName:   testDataReportNotPassFile,
		textWithPackageStatsAndFileStats: `base-package 6/8 (75%)
 first 2/4 (50%)
 second 4/4 (100%)
 third 0/0 (100%)
base-package/otherpackage 0/7 (0%)
 first 0/7 (0%)`,
		textWithPackageStats: `base-package 6/8 (75%)
base-package/otherpackage 0/7 (0%)`,
		textWithFileStats: `base-package/first 2/4 (50%)
base-package/second 4/4 (100%)
base-package/third 0/0 (100%)
base-package/otherpackage/first 0/7 (0%)`,
	},
	{
		shouldPass: true,
		fileName:   testDataReportPassFile,
		textWithPackageStatsAndFileStats: `base-package 6/6 (100%)
 first 2/2 (100%)
 second 4/4 (100%)`,
		textWithPackageStats: `base-package 6/6 (100%)`,
		textWithFileStats: `base-package/first 2/2 (100%)
base-package/second 4/4 (100%)`,
	},
}
