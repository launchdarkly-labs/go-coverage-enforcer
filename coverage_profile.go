package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// CoverageProfile represents the contents of a coverage profile created by "go test".
type CoverageProfile struct {
	// CoverageMode is one of the coverage modes supported by "go test" (default: "set"), which
	// determines how coverage counts are computed.
	CoverageMode string

	// Blocks is a list of code range items in the order that they appeared in the profile.
	Blocks []CodeBlockCoverage
}

// CodeBlockCoverage represents a single item in a coverage profile, describing the coverage of
// some section of code.
type CodeBlockCoverage struct {
	CodeRange

	// StatementCount is the number of statements in the code range.
	StatementCount int

	// CoverageCount is computed by "go test" to indicate the level of coverage for this code
	// range. If the mode was "set", it is 0 for no coverage or 1 for any coverage. If the mode
	// was "count" or "atomic", it is the number of times that "go test" detected that code
	// being executed.
	CoverageCount int
}

// CodeRange represents a section of source code.
type CodeRange struct {
	// FilePath is the path to the source file, in the format "PACKAGE_PATH/FILE_PATH" where
	// PACKAGE_PATH is the import path of the package being tested and FILE_PATH is the file's
	// relative path within the package.
	FilePath string

	// StartLine is the starting line number of the range, where 1 is the first line.
	StartLine int

	// StartColumn is the starting column number of the range, where 1 is the first column.
	StartColumn int

	// EndLine is the ending line number of the range, where 1 is the first line.
	EndLine int

	// EndColumn is the ending column number of the range, where 1 is the first column.
	EndColumn int
}

// GetPackagePathAndFileName parses a file path like "github.com/a/b/c/d.go" into "github.com/a/b/c"
// and "d.go".
func (r CodeRange) GetPackagePathAndFileName() (string, string) {
	p := r.FilePath[:strings.LastIndex(r.FilePath, "/")]
	return p, strings.TrimPrefix(r.FilePath, p+"/")
}

// ReadCoverageProfile attempts to parse a coverage profile generated by "go test".
//
// The standard format of this file is a first line "mode: X" where X is one of the coverage modes
// supported by "go test" (default: "set"), which determines how coverage counts are computed,
// followed by any number of lines which specify a source code range, the number of statements
// that were compiled in that range, and the computed coverage count.
func ReadCoverageProfile(reader io.Reader) (*CoverageProfile, error) {
	ret := &CoverageProfile{}
	coverageLineRegex := regexp.MustCompile(`^(.*):(\d+)\.(\d+),(\d+)\.(\d+) +(\d+) +(\d+)$`)

	scanner := bufio.NewScanner(reader)
	n := 0
	for scanner.Scan() {
		n++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "mode:") {
			ret.CoverageMode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
		} else {
			if matches := coverageLineRegex.FindStringSubmatch(line); matches != nil {
				var span CodeBlockCoverage
				span.CodeRange.FilePath = matches[1]
				span.CodeRange.StartLine, _ = strconv.Atoi(matches[2])
				span.CodeRange.StartColumn, _ = strconv.Atoi(matches[3])
				span.CodeRange.EndLine, _ = strconv.Atoi(matches[4])
				span.CodeRange.EndColumn, _ = strconv.Atoi(matches[5])
				span.StatementCount, _ = strconv.Atoi(matches[6])
				span.CoverageCount, _ = strconv.Atoi(matches[7])
				ret.Blocks = append(ret.Blocks, span)
			} else {
				return nil, fmt.Errorf("Invalid profile data format at line %d", n)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

// WriteTo writes the profile data to a Writer in the same format that it was parsed from.
func (cp CoverageProfile) WriteTo(writer io.Writer) error {
	bw := bufio.NewWriter(writer)
	_, err := bw.WriteString(fmt.Sprintf("mode: %s\n", cp.CoverageMode))
	if err != nil {
		// COVERAGE: there is no way to simulate this condition in unit tests
		return err
	}
	for _, b := range cp.Blocks {
		line := fmt.Sprintf(
			"%s:%d.%d,%d.%d %d %d\n",
			b.CodeRange.FilePath,
			b.CodeRange.StartLine,
			b.CodeRange.StartColumn,
			b.CodeRange.EndLine,
			b.CodeRange.EndColumn,
			b.StatementCount,
			b.CoverageCount,
		)
		_, err = bw.WriteString(line)
		if err != nil {
			// COVERAGE: there is no way to simulate this condition in unit tests
			return err
		}
	}
	return bw.Flush()
}

// GetUniqueBlocks returns a sorted, deduplicated slice of profile items.
//
// The profile generated by "go test" can include many lines for the same code range, as it iterates
// through the covered code paths. If any of those lines has a nonzero coverage count, then that code
// range was covered. This function reduces any number of items that reference the same code range to
// a single item that has a nonzero coverage count if any such count was present, or zero otherwise.
// It also sorts items in ascending order of package path, file path, and starting line number.
func (cp CoverageProfile) GetUniqueBlocks() []CodeBlockCoverage {
	workMap := make(map[CodeRange]CodeBlockCoverage)
	for _, b := range cp.Blocks {
		if b.CoverageCount > 0 {
			workMap[b.CodeRange] = b
		} else {
			if _, ok := workMap[b.CodeRange]; !ok {
				workMap[b.CodeRange] = b
			}
		}
	}

	var ret []CodeBlockCoverage
	for _, b := range workMap {
		ret = append(ret, b)
	}
	sort.Slice(ret, func(i, j int) bool {
		b0 := ret[i]
		b1 := ret[j]
		pp0, fn0 := b0.CodeRange.GetPackagePathAndFileName()
		pp1, fn1 := b1.CodeRange.GetPackagePathAndFileName()
		if pp0 != pp1 {
			return pp0 < pp1
		}
		if fn0 != fn1 {
			return fn0 < fn1
		}
		return b0.CodeRange.StartLine < b1.CodeRange.StartLine
	})
	return ret
}

// WithBlockFilter creates a copy of this CoverageProfile, removing any blocks for which the retainBlock
// function returns false.
func (cp CoverageProfile) WithBlockFilter(retainBlock func(CodeBlockCoverage) bool) *CoverageProfile {
	ret := new(CoverageProfile)
	*ret = cp
	ret.Blocks = make([]CodeBlockCoverage, 0, len(cp.Blocks))
	for _, b := range cp.Blocks {
		if retainBlock(b) {
			ret.Blocks = append(ret.Blocks, b)
		}
	}
	return ret
}
