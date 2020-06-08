package main

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// InferPackagePath attempts to determine the import path of the package in the current directory -
// preferably from go.mod, otherwise from git. Returns "" if unsuccessful.
func InferPackagePath() string {
	if goModFile, err := os.Open("go.mod"); err == nil {
		defer goModFile.Close()
		scanner := bufio.NewScanner(goModFile)
		if scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "module ") {
				return strings.TrimSpace(strings.TrimPrefix(line, "module "))
			}
		}
	}
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		gitURL := strings.TrimSpace(string(out))
		matches := regexp.MustCompile(`^git@(.*):(.*)\.git$`).FindStringSubmatch(gitURL)
		if len(matches) == 3 {
			return matches[1] + "/" + matches[2]
		}
		matches = regexp.MustCompile(`^https?://(.*)\.git$`).FindStringSubmatch(gitURL)
		if len(matches) == 2 {
			// COVERAGE: there is no way to simulate this condition in unit tests, because git changes
			// the HTTPS URL to SSH
			return matches[1]
		}
		matches = regexp.MustCompile(`^ssh://git@(.*)\.git$`).FindStringSubmatch(gitURL)
		if len(matches) == 2 {
			return matches[1]
		}
	}

	return ""
}
