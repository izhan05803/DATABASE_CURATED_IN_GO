.PHONY: build test test-race bench lint run clean test-one coverage

# Build the binary
build:
	go build -o godb.exe ./cmd/godb

# Run all tests (without race detector for Windows compatibility)
test:
	go test ./...

# Run tests with race detector (requires CGO_ENABLED=1)
test-race:
	go test -race ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run linting
lint:
	go fmt ./...
	go vet ./...

# Build and run
run: build
	./godb.exe

# Clean build artifacts
clean:
	del /Q godb.exe 2>nul || true
	del /Q data\*.godb 2>nul || true

# Run a specific test
# Usage: make test-one TEST=TestBTreeInsert
test-one:
	go test -v -run $(TEST) ./...

# Check test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
