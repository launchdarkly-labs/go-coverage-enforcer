# Go test coverage enforcer

This command-line tool can be used in conjunction with Go's built-in coverage profiling to provide CI enforcement of coverage goals.

Unlike other tools for a similar purpose, `go-coverage-enforcer` does not generate the coverage profile itself, does not instrument the code, and has no opinions about any particular patterns in the code. It is simply a post-processing step after a coverage profile has already been generated with `go test`.

## Usage

```shell
go install github.com/eli-darkly/go-coverage-enforcer

go test . -coverprofile coverage.out
go-coverage-enforcer coverage.out
```

The working directory should be the base directory of the package that is being tested.

The command writes results to standard output, and returns exit code 1 if any problems were found, 0 otherwise.

## Options

**`-package IMPORTPATH`**

Specifies the import path of the current package that is being tested.

If you don't specify this, `go-coverage-enforcer` will try to determine the package by looking for a `go.mod` file in the current directory. As a fallback, it will check whether the current directory is a Git checkout and will try to determine the import path from the Git URL.

**`-showcode`**

Causes the output to include the source code of each uncovered block.

With `-showcode`:

```
Uncovered blocks detected:

somepackage/some_file.go 133-135
133>        if n == 0 {
134>            return nil
135>        }
```

Without `-showcode`:
```
Uncovered blocks detected:

somepackage/some_file.go 133-135
```

**`-skipfiles PATTERN`**

If provided, this must be a valid regular expression. Any files whose relative path matches this expression will be skipped when analyzing the coverage profile.

**`-skipcode PATTERN`**

If provided, this must be a valid regular expression. Uncovered code blocks containing any line matching this expression will be skipped.

This is a way to add explicit overrides of coverage checking in code paths where test coverage is impossible. For instance, specifying `-skipcode "// NOCOVER" would allow usage like this:

```go
    if err != nil {
        return err // NOCOVER: there is no way to cause this error in unit tests
    }
```

This only affects the processing done by `go-coverage-enforcer`-- not the original coverage report generated by `go test`.

**`-outprofile FILEPATH`**

This causes `go-coverage-enforcer` to write the profile data to the specified path, in the same format that was generated by `go test`, after removing any code blocks that were skipped due to `-skipfiles` or `-skipcode`.

You can then run `go cover` on the filtered file to generate coverage reports that reflect this filtering. For instance, if you run `go cover -html=FILEPATH` to view the report as a web page, skipped files will not appear and skipped code blocks will appear in gray rather than red or green, and coverage percentages will be calculated as if the skipped files and blocks did not exist.
