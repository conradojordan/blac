.DEFAULT_GOAL := run

VERSION := 1.0.0
LINUX_AMD64 := blac-$(VERSION)-amd64-linux
MACOS_AMD64 := blac-$(VERSION)-amd64-apple-darwin
WINDOWS_AMD64 := blac-$(VERSION)-amd64-windows
MACOS_ARM64 := blac-$(VERSION)-arm64-apple-darwin

build:
	@echo "Compiling Blac v$(VERSION) for Linux (amd64), MacOS (amd64 and arm64) and Windows (amd64)"
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(LINUX_AMD64) .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(MACOS_AMD64) .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(MACOS_ARM64) .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(WINDOWS_AMD64) .

run:
	@echo "Running Blac..."
	@go run .

clean:
	@echo "Deleting compiled binaries: $(LINUX_AMD64), $(MACOS_AMD64), $(MACOS_ARM64), $(WINDOWS_AMD64)"
	@rm -f $(LINUX_AMD64)
	@rm -f $(MACOS_AMD64)
	@rm -f $(MACOS_ARM64)
	@rm -f $(WINDOWS_AMD64)

