# ====================================================================================
# PROJECT AUTOMATION (Go translation of package.json scripts)
# ====================================================================================

.PHONY: all clear prebuild build postbuild test script lint codeclimate help

# Variables
BINARY_NAME=go-csv
DIST_DIR=dist
COVERAGE_DIR=test/coverage

all: clear lint test build

## clear: Remove build modules, build folders, test reports, and temporary files (replaces rimraf)
clear:
	rm -rf $(DIST_DIR) test/
	rm -f junit.xml coverage.out

## prebuild: Enforce styling corrections before compilation
prebuild:
	go fmt ./...
	/c/bin/golangci-lint run --fix

## build: Compile the binary executable production build
build: prebuild script
	@echo "Build complete."

## postbuild: Run automated formatting audits across the repository
postbuild:
	/c/bin/golangci-lint run

## test: Execute all module tests across all folders, record code coverage metrics, and print reports
test:
	mkdir -p $(COVERAGE_DIR)
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o $(COVERAGE_DIR)/index.html
	@echo "Coverage report written to $(COVERAGE_DIR)/index.html"

## script: Compile source package tools using your root main entry point if available
script:
	mkdir -p $(DIST_DIR)
	@if [ -f main.go ]; then \
		go build -v -o $(DIST_DIR)/$(BINARY_NAME).exe main.go; \
	else \
		go build -v -o $(DIST_DIR)/$(BINARY_NAME).dll ./pkg/csv; \
		echo "No main.go found. Compiled package into reusable distribution library instead."; \
	fi

## lint: Run strict golangci-lint quality configurations
lint:
	/c/bin/golangci-lint run

## codeclimate: Replicates package.json dockerized static infrastructure analyzer block
codeclimate:
	docker run --interactive --rm --env CODECLIMATE_CODE="$$PWD" --volume "$$PWD":/code --volume /var/run/docker.sock:/var/run/docker.sock --volume /tmp/cc:/tmp/cc codeclimate/codeclimate analyze

## help: Output operational manual descriptions for shortcuts
help:
	@echo "Available tasks:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
