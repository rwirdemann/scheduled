VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.version=$(VERSION)

UNAME_OS := $(shell uname -s)
ifeq ($(UNAME_OS),Darwin)
SHA256 := shasum -a 256
else
SHA256 := sha256sum
endif

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/scheduled-darwin-amd64 cmd/scheduled.go

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/scheduled-darwin-arm64 cmd/scheduled.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/scheduled-linux-amd64 cmd/scheduled.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/scheduled-linux-arm64 cmd/scheduled.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/scheduled-windows-amd64.exe cmd/scheduled.go

build-all: test build-darwin build-darwin-arm64 build-linux build-linux-arm64 build-windows checksums
	@echo "Built binaries for Darwin, Linux and Windows in bin/"

checksums:
	@echo "Generating SHA256 checksums..."
	@$(SHA256) bin/scheduled-darwin-amd64 | awk '{print $$1}' > bin/scheduled-darwin-amd64.sha256
	@$(SHA256) bin/scheduled-darwin-arm64 | awk '{print $$1}' > bin/scheduled-darwin-arm64.sha256
	@$(SHA256) bin/scheduled-linux-amd64 | awk '{print $$1}' > bin/scheduled-linux-amd64.sha256
	@$(SHA256) bin/scheduled-linux-arm64 | awk '{print $$1}' > bin/scheduled-linux-arm64.sha256
	@$(SHA256) bin/scheduled-windows-amd64.exe | awk '{print $$1}' > bin/scheduled-windows-amd64.exe.sha256
	@echo "Checksums written to bin/*.sha256"

test:
	go test ./...

install: build-all
ifeq ($(shell uname),Darwin)
ifeq ($(shell uname -m),arm64)
	@echo "Installing darwin ARM64 binary to ${GOROOT}/bin/scheduled"
	@cp bin/scheduled-darwin-arm64 ${GOROOT}/bin/scheduled
else
	@echo "Installing darwin AMD64 binary to ${GOROOT}/bin/scheduled"
	@cp bin/scheduled-darwin-amd64 ${GOROOT}/bin/scheduled
endif
else ifeq ($(shell uname),Linux)
ifeq ($(shell uname -m),aarch64)
	@echo "Installing Linux ARM64 binary to ${GOROOT}/bin/scheduled"
	@cp bin/scheduled-linux-arm64 ${GOROOT}/bin/scheduled
else
	@echo "Installing Linux AMD64 binary to ${GOROOT}/bin/scheduled"
	@cp bin/scheduled-linux-amd64 ${GOROOT}/bin/scheduled
endif
else
	@echo "Installing Windows binary to ${GOROOT}/bin/scheduled.exe"
	@cp bin/scheduled-windows-amd64.exe ${GOROOT}/bin/scheduled.exe
endif

.PHONY: build-darwin build-darwin-arm64 build-linux build-linux-arm64 build-windows build-all checksums test install
