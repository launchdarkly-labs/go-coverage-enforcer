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

func TestReadCoverageProfile(t *testing.T) {
	t.Run("parses file with expected values", func(t *testing.T) {
		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			assert.Equal(t, "set", cp.CoverageMode)
			assert.Equal(t, expectedParsedCoverageProfile.Blocks, cp.Blocks)
		})
	})

	t.Run("fails for reader error", func(t *testing.T) {
		r := &mockReaderWriterThatReturnsError{err: errors.New("sorry")}
		_, err := ReadCoverageProfile(r)
		assert.Equal(t, r.err, err)
	})

	t.Run("fails for malformed data", func(t *testing.T) {
		withTestProfile("coverage_data_malformed", func(cp *CoverageProfile, err error) {
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid profile data format")
		})
	})
}

func TestCoverageProfileGetUniqueBlocks(t *testing.T) {
	withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
		blocks := cp.GetUniqueBlocks()
		expected := []CodeBlockCoverage{
			// the blocks are sorted in order of package path, file path, and starting line, and
			// duplicates are filtered out
			expectedParsedCoverageProfile.Blocks[1], // first 1.1,2.1
			expectedParsedCoverageProfile.Blocks[0], // first 4.4,5.2
			expectedParsedCoverageProfile.Blocks[3], // second 1.1,5.1
			expectedParsedCoverageProfile.Blocks[4], // third 1.1,2.1
			expectedParsedCoverageProfile.Blocks[6], // third 3.1,4.1
			expectedParsedCoverageProfile.Blocks[7], // third 4.1,5.1
			expectedParsedCoverageProfile.Blocks[2], // otherpackage/first 2.1,2.10
		}
		assert.Equal(t, expected, blocks)
	})
}

func TestCoverageProfileWriteTo(t *testing.T) {
	t.Run("output matches input", func(t *testing.T) {
		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			buf := new(bytes.Buffer)
			err := cp.WriteTo(buf)
			require.NoError(t, err)

			trimmedLines := func(data []byte) []string {
				return strings.Split(strings.TrimSpace(string(data)), "\n")
			}

			expectedData, _ := ioutil.ReadFile(testDataMainFile)

			assert.Equal(t, trimmedLines(expectedData), trimmedLines(buf.Bytes()))
		})
	})

	t.Run("fails for writer error", func(t *testing.T) {
		withValidTestProfile(testDataMainFile, func(cp *CoverageProfile) {
			w := &mockReaderWriterThatReturnsError{err: errors.New("sorry")}
			err := cp.WriteTo(w)
			require.Equal(t, w.err, err)
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
