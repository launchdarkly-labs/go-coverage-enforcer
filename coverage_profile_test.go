package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCoverageProfileFileName = "test-coverage-profile"

var expectedParsedCoverageProfile = CoverageProfile{ // the parsed content of the file above
	CoverageMode: "set",
	Blocks: []CodeBlockCoverage{ // the parsed content of the file above
		{CodeRange{"github.com/launchdarkly/example/file1", 20, 2, 30, 12}, 1, 0},
		{CodeRange{"github.com/launchdarkly/example/package1/file2", 25, 37, 26, 15}, 1, 0},
		{CodeRange{"github.com/launchdarkly/example/package1/file3", 97, 47, 99, 2}, 1, 0},
		{CodeRange{"github.com/launchdarkly/example/package1/file3", 97, 47, 99, 2}, 1, 0},
		{CodeRange{"github.com/launchdarkly/example/package1/file2", 35, 12, 36, 16}, 1, 0},
		{CodeRange{"github.com/launchdarkly/example/package1/file2", 35, 12, 36, 16}, 1, 1},
		{CodeRange{"github.com/launchdarkly/example/package1/file3", 97, 47, 99, 2}, 1, 1},
		{CodeRange{"github.com/launchdarkly/example/file1", 10, 42, 10, 52}, 1, 0},
	},
}

func TestReadCoverageProfile(t *testing.T) {
	t.Run("parses file with expected values", func(t *testing.T) {
		withValidTestProfile(testCoverageProfileFileName, func(cp *CoverageProfile) {
			assert.Equal(t, "set", cp.CoverageMode)
			assert.Equal(t, expectedParsedCoverageProfile.Blocks, cp.Blocks)
		})
	})

	t.Run("skips blank lines", func(t *testing.T) {
		withValidTestProfile("test-coverage-profile-with-blank-line", func(cp *CoverageProfile) {
			assert.Equal(t, expectedParsedCoverageProfile, *cp)
		})
	})

	t.Run("fails for reader error", func(t *testing.T) {
		r := &mockReaderWriterThatReturnsError{err: errors.New("sorry")}
		_, err := ReadCoverageProfile(r)
		assert.Equal(t, r.err, err)
	})

	t.Run("fails for malformed data", func(t *testing.T) {
		withTestProfile("test-coverage-profile-malformed-data", func(cp *CoverageProfile, err error) {
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid profile data format")
		})
	})
}

func TestCoverageProfileMethods(t *testing.T) {
	t.Run("GetUniqueBlocks", func(t *testing.T) {
		withValidTestProfile(testCoverageProfileFileName, func(cp *CoverageProfile) {
			blocks := cp.GetUniqueBlocks()
			expected := []CodeBlockCoverage{
				// the blocks are sorted in order of file path and starting line, and duplicates are filtered out
				expectedParsedCoverageProfile.Blocks[7],
				expectedParsedCoverageProfile.Blocks[0],
				expectedParsedCoverageProfile.Blocks[1],
				expectedParsedCoverageProfile.Blocks[5],
				expectedParsedCoverageProfile.Blocks[6],
			}
			assert.Equal(t, expected, blocks)
		})
	})

	t.Run("WriteTo", func(t *testing.T) {
		t.Run("output matches input", func(t *testing.T) {
			withValidTestProfile(testCoverageProfileFileName, func(cp *CoverageProfile) {
				buf := new(bytes.Buffer)
				err := cp.WriteTo(buf)
				require.NoError(t, err)

				expectedData, _ := ioutil.ReadFile(testCoverageProfileFileName)
				assert.Equal(t, strings.Split(string(expectedData), "\n"), strings.Split(buf.String(), "\n"))
			})
		})

		t.Run("fails for writer error", func(t *testing.T) {
			withValidTestProfile(testCoverageProfileFileName, func(cp *CoverageProfile) {
				w := &mockReaderWriterThatReturnsError{err: errors.New("sorry")}
				err := cp.WriteTo(w)
				require.Equal(t, w.err, err)
			})
		})
	})
}

type mockReaderWriterThatReturnsError struct {
	err error
}

func (m *mockReaderWriterThatReturnsError) Read(p []byte) (int, error) {
	return 0, m.err
}

func (m *mockReaderWriterThatReturnsError) Write(data []byte) (int, error) {
	return 0, m.err
}
