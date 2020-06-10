package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withTempDir(action func(dirPath string)) {
	path, err := ioutil.TempDir("", "go-coverage-enforcer-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(path)
	action(path)
}

func TestInferPackagePathFromGoMod(t *testing.T) {
	assert.Equal(t, "github.com/launchdarkly-labs/go-coverage-enforcer", InferPackagePath())
}

func TestInferPackagePathFromGitURLWithSSH(t *testing.T) {
	withTempDir(func(dirPath string) {
		inWorkingDir(dirPath, func() {
			require.NoError(t, exec.Command("git", "init").Run())
			require.NoError(t, exec.Command("git", "remote", "add", "origin", "git@github.com:fake/path.git").Run())

			assert.Equal(t, "github.com/fake/path", InferPackagePath())
		})
	})
}

func TestInferPackagePathFromGitURLWithHTTPS(t *testing.T) {
	withTempDir(func(dirPath string) {
		inWorkingDir(dirPath, func() {
			require.NoError(t, exec.Command("git", "init").Run())
			require.NoError(t, exec.Command("git", "remote", "add", "origin", "https://github.com/fake/path.git").Run())

			assert.Equal(t, "github.com/fake/path", InferPackagePath())
		})
	})
}
func TestInferPackagePathFails(t *testing.T) {
	withTempDir(func(dirPath string) {
		inWorkingDir(dirPath, func() {
			assert.Equal(t, "", InferPackagePath())
		})
	})
}
