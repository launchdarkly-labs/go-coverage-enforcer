package main

import "os"

const testDataDir = "./testdata"

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
