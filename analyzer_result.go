package main

// AnalyzerResult is the result returned by AnalyzeCoverage.
type AnalyzerResult struct {
	// Package is a list of the analyzed packages. It does not include packages whose files were
	// all skipped with "-skipfiles".
	Packages []AnalyzerPackageResult

	// SkippedFilePaths is a list of file paths that were skipped due to the "-skipfiles" option.
	// These are in the same format as in the coverage profile, so they include the package's
	// import path.
	SkippedFilePaths []string

	// SkippedBlocks is a list of code ranges that were skipped due to the "-skipcode" option.
	SkippedBlocks []CodeBlockCoverage
}

// AnalyzerPackageResult is package-level information in AnalyzerResult.
type AnalyzerPackageResult struct {
	// RelativePath is the package's path within the main package. If the main package is
	// "github.com/example/package", then the relative path of that package is "", and the
	// relative path of "github.com/example/package/a/b" is "a/b".
	RelativePath string

	// Files is a list of the analyzed files in this package. It does not include files that were
	// skipped with "-skipfiles".
	Files []AnalyzerFileResult
}

// AnalyzerFileResult is file-level information in AnalyzerResult.
type AnalyzerFileResult struct {
	// FileName is the simple filename, without the package path.
	FileName string

	// TotalStatements is the number of statements in all blocks in this file that were
	// included in the coverage profile.
	TotalStatements int

	// CoveredStatements is the number of statements in all blocks in this file that were
	// reported as covered in the coverage profile.
	CoveredStatements int

	// UncoveredBlocks are the code blocks in this file that lacked coverage. The list is sorted in
	// ascending order of starting line number. It does not include any locations that were
	// skipped with "-skipcode".
	UncoveredBlocks []UncoveredBlock
}

// UncoveredBlock is a code range that had no coverage.
type UncoveredBlock struct {
	CodeRange

	// Text is an optional excerpt of the source code file corresponding to the range's starting
	// and ending line numbers. It is only provided if the "-showcode" option was used; otherwise
	// it is nil.
	Text []string
}
