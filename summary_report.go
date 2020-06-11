package main

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// SummaryReport is a post-processed version of the data from AnalyzerResult, corresponding to how we will
// display the information.
type SummaryReport struct {
	Packages        []SummaryReportPackage
	UncoveredBlocks []UncoveredBlock
	Pass            bool
}

type SummaryReportPackage struct {
	FullPackagePath string
	Files           []SummaryReportFile
	Coverage        SummaryReportCoverage
}

type SummaryReportFile struct {
	FileName string
	Coverage SummaryReportCoverage
}

type SummaryReportCoverage struct {
	TotalStatements   int
	CoveredStatements int
}

func (c SummaryReportCoverage) GetCoveredPercent() int {
	if c.TotalStatements == 0 {
		return 100
	}
	return c.CoveredStatements * 100 / c.TotalStatements
}

func NewSummaryReport(result AnalyzerResult, opts EnforcerOptions) SummaryReport {
	var r SummaryReport
	for _, p := range result.Packages {
		var rp SummaryReportPackage
		rp.FullPackagePath = opts.PackagePath
		if p.RelativePath != "" {
			rp.FullPackagePath += "/" + p.RelativePath
		}
		for _, f := range p.Files {
			rp.Coverage.TotalStatements += f.TotalStatements
			rp.Coverage.CoveredStatements += f.CoveredStatements
			rp.Files = append(rp.Files, SummaryReportFile{
				FileName: f.FileName,
				Coverage: SummaryReportCoverage{
					TotalStatements: f.TotalStatements, CoveredStatements: f.CoveredStatements},
			})
			r.UncoveredBlocks = append(r.UncoveredBlocks, f.UncoveredBlocks...)
		}
		r.Packages = append(r.Packages, rp)
	}
	sort.Slice(r.Packages, func(i, j int) bool {
		return r.Packages[i].FullPackagePath < r.Packages[j].FullPackagePath
	})
	r.Pass = len(r.UncoveredBlocks) == 0
	return r
}

func (r SummaryReport) Output(writer io.Writer, opts EnforcerOptions) bool {
	if opts.ShowPackageStats || opts.ShowFileStats {
		tw := tabwriter.NewWriter(writer, 1, 4, 1, ' ', 0)
		for _, p := range r.Packages {
			if opts.ShowPackageStats {
				fmt.Fprintf(tw, "%s\t%d/%d\t(%d%%)\t\n",
					p.FullPackagePath,
					p.Coverage.CoveredStatements,
					p.Coverage.TotalStatements,
					p.Coverage.GetCoveredPercent(),
				)
			}
			if opts.ShowFileStats {
				for _, f := range p.Files {
					var desc string
					if opts.ShowPackageStats {
						desc = "  " + f.FileName
					} else {
						desc = p.FullPackagePath + "/" + f.FileName
					}
					fmt.Fprintf(tw, "%s\t%d/%d\t(%d%%)\t\n",
						desc,
						f.Coverage.CoveredStatements,
						f.Coverage.TotalStatements,
						f.Coverage.GetCoveredPercent(),
					)
				}
			}
		}
		tw.Flush()
	}

	if opts.ShowPackageStats || opts.ShowFileStats {
		fmt.Fprintln(writer)
	}

	if r.Pass {
		fmt.Fprintln(writer, "Coverage scan passes!")
		return true
	}

	fmt.Fprintln(writer, "Uncovered blocks detected:")
	for _, b := range r.UncoveredBlocks {
		if opts.ShowCode {
			fmt.Fprintln(writer)
		}
		fmt.Fprintf(writer, "%s %d-%d\n", b.CodeRange.FilePath, b.CodeRange.StartLine, b.CodeRange.EndLine)
		if opts.ShowCode {
			for i, line := range b.Text {
				fmt.Fprintf(writer, "%d>\t%s\n", b.CodeRange.StartLine+i, line)
			}
		}
	}

	return false
}
