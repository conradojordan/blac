.DEFAULT_GOAL := run

LINUX_AMD64 := blac-linux-amd64
MACOS_AMD64 := blac-mac-amd64
MACOS_ARM64 := blac-mac-arm64
WINDOWS_AMD64 := blac-windows-amd64

build:
	@echo "Compiling for Linux (amd64), Mac (amd64 and arm64) and Windows (amd64)"
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(LINUX_AMD64) .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(MACOS_AMD64) .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(MACOS_ARM64) .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(WINDOWS_AMD64) .

run:
	@echo "Running project..."
	@go run .

clean:
	@echo "Deleting old binaries ($(LINUX_AMD64), $(MACOS_AMD64), $(MACOS_ARM64), $(WINDOWS_AMD64))..."
	@rm -f $(LINUX_AMD64)
	@rm -f $(MACOS_AMD64)
	@rm -f $(MACOS_ARM64)
	@rm -f $(WINDOWS_AMD64)

