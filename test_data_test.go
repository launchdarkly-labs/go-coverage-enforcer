package main

import "os"

const (
	testDataDir         = "./testdata"
	testDataMainFile    = "coverage_data_for_basic_tests"
	testDataPackagePath = "base-package"
)

var testBaseOptions = EnforcerOptions{
	PackagePath: testDataPackagePath,
}

func inWorkingDir(path string, action func()) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer os.Chdir(dir)
	err = os.Chdir(path)
	if err != nil {
		panic(err)
	}
	action()
}

func inTestDataDir(action func()) {
	inWorkingDir(testDataDir, action)
}

func withTestProfile(filename string, action func(*CoverageProfile, error)) {
	inTestDataDir(func() {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		cp, err := ReadCoverageProfile(f)
		action(cp, err)
	})
}

func withValidTestProfile(filename string, action func(*CoverageProfile)) {
	withTestProfile(filename, func(cp *CoverageProfile, err error) {
		if err != nil {
			panic(err)
		}
		action(cp)
	})
}

// The parsed content of "coverage_data_for_basic_tests" - exactly as it is, with no post-processing
var expectedParsedCoverageProfile = CoverageProfile{
	CoverageMode: "set",
	Blocks: []CodeBlockCoverage{ // the parsed content of the file above
		{CodeRange{"base-package/first", 3, 4, 5, 2}, 1, 0},
		{CodeRange{"base-package/first", 1, 1, 2, 1}, 2, 0},
		{CodeRange{"base-package/otherpackage/first", 2, 1, 2, 10}, 1, 0},
		{CodeRange{"base-package/second", 1, 1, 5, 1}, 5, 0},
		{CodeRange{"base-package/third", 1, 1, 2, 1}, 2, 0},
		{CodeRange{"base-package/second", 1, 1, 5, 1}, 5, 0},
		{CodeRange{"base-package/third", 3, 1, 4, 1}, 2, 1},
		{CodeRange{"base-package/third", 4, 1, 5, 1}, 2, 0},
	},
}

func makeTestAnalyzerExpectedResult() AnalyzerResult {
	r := makeTestAnalyzerExpectedResultWithCode()
	for i, p := range r.Packages {
		for j, f := range p.Files {
			for k := range f.UncoveredBlocks {
				r.Packages[i].Files[j].UncoveredBlocks[k].Text = nil
			}
		}
	}
	return r
}

func makeTestAnalyzerExpectedResultWithCode() AnalyzerResult {
	p1f1 := AnalyzerFileResult{FileName: "first", TotalStatements: 3, CoveredStatements: 0}
	p1f1.UncoveredBlocks = append(p1f1.UncoveredBlocks, UncoveredBlock{
		CodeRange: CodeRange{FilePath: testDataPackagePath + "/first",
			StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1},
		Text: []string{"first file line 1", "first file line 2"},
	})
	p1f1.UncoveredBlocks = append(p1f1.UncoveredBlocks, UncoveredBlock{
		CodeRange: CodeRange{FilePath: testDataPackagePath + "/first",
			StartLine: 3, StartColumn: 4, EndLine: 5, EndColumn: 2},
		Text: []string{"first file line 3", "first file line 4", "first file line 5"},
	})

	p1f2 := AnalyzerFileResult{FileName: "second", TotalStatements: 5, CoveredStatements: 0}
	p1f2.UncoveredBlocks = append(p1f2.UncoveredBlocks, UncoveredBlock{
		CodeRange: CodeRange{FilePath: testDataPackagePath + "/second",
			StartLine: 1, StartColumn: 1, EndLine: 5, EndColumn: 1},
		Text: []string{"second file line 1", "second file line 2", "second file line 3", "second file line 4", "second file line 5"},
	})

	p1f3 := AnalyzerFileResult{FileName: "third", TotalStatements: 6, CoveredStatements: 2}
	p1f3.UncoveredBlocks = append(p1f3.UncoveredBlocks, UncoveredBlock{
		CodeRange: CodeRange{FilePath: testDataPackagePath + "/third",
			StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1},
		Text: []string{"third file line 1", "third file line 2"},
	})
	p1f3.UncoveredBlocks = append(p1f3.UncoveredBlocks, UncoveredBlock{
		CodeRange: CodeRange{FilePath: testDataPackagePath + "/third",
			StartLine: 4, StartColumn: 1, EndLine: 5, EndColumn: 1},
		Text: []string{"third file line 4", "third file line 5"},
	})

	p1 := AnalyzerPackageResult{RelativePath: "", Files: []AnalyzerFileResult{p1f1, p1f2, p1f3}}

	p2f1 := AnalyzerFileResult{FileName: "first", TotalStatements: 1, CoveredStatements: 0}
	p2f1.UncoveredBlocks = append(p2f1.UncoveredBlocks, UncoveredBlock{
		CodeRange: CodeRange{FilePath: testDataPackagePath + "/otherpackage/first",
			StartLine: 2, StartColumn: 1, EndLine: 2, EndColumn: 10},
		Text: []string{"other package first file line 2"},
	})

	p2 := AnalyzerPackageResult{RelativePath: "otherpackage", Files: []AnalyzerFileResult{p2f1}}

	return AnalyzerResult{
		Packages: []AnalyzerPackageResult{p1, p2},
	}
}
