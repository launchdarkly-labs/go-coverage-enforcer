
EXECUTABLE=./go-coverage-enforcer
COVERAGE_PROFILE=coverage.out
ALL_SOURCES := $(shell find * -type f -name "*.go")

.PHONY: build test self-coverage

$(EXECUTABLE): *.go
	go build

build: $(EXECUTABLE)

$(COVERAGE_PROFILE): $(ALL_SOURCES)
	go test . -coverprofile=$(COVERAGE_PROFILE)

test: $(COVERAGE_PROFILE)

self-coverage: $(EXECUTABLE) $(COVERAGE_PROFILE)
	$(EXECUTABLE) -showcode -packagestats -filestats -skipfiles main.go -skipcode '// COVERAGE' $(COVERAGE_PROFILE)
