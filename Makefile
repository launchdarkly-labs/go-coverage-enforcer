
EXECUTABLE=./go-coverage-enforcer
COVERAGE_PROFILE=coverage.out

.PHONY: build test self-coverage

$(EXECUTABLE): *.go
	go build

build: $(EXECUTABLE)

$(COVERAGE_PROFILE):
	go test . -coverprofile=$(COVERAGE_PROFILE)

test: $(COVERAGE_PROFILE)

self-coverage: $(EXECUTABLE) $(COVERAGE_PROFILE)
	$(EXECUTABLE) -showcode -skipfiles main.go -skipcode '// COVERAGE' $(COVERAGE_PROFILE)
