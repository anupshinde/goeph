.PHONY: build build-examples test test-v cover cover-html vet clean

# Build all packages (excluding examples)
build:
	go build ./spk/ ./coord/ ./timescale/ ./satellite/ ./star/ ./lunarnodes/ ./units/ ./geometry/ ./elements/ ./magnitude/

# Build examples (compile check only)
build-examples:
	go build ./examples/...

# Run all tests
test:
	go test ./spk/ ./coord/ ./timescale/ ./satellite/ ./star/ ./lunarnodes/ ./units/ ./geometry/ ./elements/ ./magnitude/

# Run all tests with verbose output
test-v:
	go test -v ./spk/ ./coord/ ./timescale/ ./satellite/ ./star/ ./lunarnodes/ ./units/ ./geometry/ ./elements/ ./magnitude/

# Run tests with coverage and print summary
cover:
	go test -coverprofile=coverage.out ./spk/ ./coord/ ./timescale/ ./satellite/ ./star/ ./lunarnodes/ ./units/ ./geometry/ ./elements/ ./magnitude/
	go tool cover -func=coverage.out

# Generate HTML coverage report
cover-html: cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run go vet
vet:
	go vet ./spk/ ./coord/ ./timescale/ ./satellite/ ./star/ ./lunarnodes/ ./units/ ./geometry/ ./elements/ ./magnitude/

# Run a single test by name: make test-one TEST=TestObserveGolden PKG=./spk/
test-one:
	go test -v -run $(TEST) $(PKG)

# Clean generated files
clean:
	rm -f coverage.out coverage.html
