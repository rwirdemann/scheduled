build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/scheduled-macos-amd64 cmd/scheduled.go

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/scheduled-macos-arm64 cmd/scheduled.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/scheduled-linux-amd64 cmd/scheduled.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/scheduled-linux-arm64 cmd/scheduled.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/scheduled-windows-amd64.exe cmd/scheduled.go

build-all: build-darwin build-darwin-arm64 build-linux build-windows
	@echo "Built binaries for Darwin, Linux and Windows in bin/"

install: build-all
ifeq ($(shell uname),Darwin)
ifeq ($(shell uname -m),arm64)
	@echo "Installing macOS ARM64 binary to ${GOPATH}/bin/scheduled"
	@cp bin/scheduled-macos-arm64 ${GOPATH}/bin/scheduled
else
	@echo "Installing macOS AMD64 binary to ${GOPATH}/bin/scheduled"
	@cp bin/scheduled-macos-amd64 ${GOPATH}/bin/scheduled
endif
else ifeq ($(shell uname),Linux)
	@echo "Installing Linux binary to ${GOPATH}/bin/scheduled"
	@cp bin/scheduled-linux-amd64 ${GOPATH}/bin/scheduled
else
	@echo "Installing Windows binary to ${GOPATH}/bin/scheduled.exe"
	@cp bin/scheduled-windows-amd64.exe ${GOPATH}/bin/scheduled.exe
endif

.PHONY: build-darwin build-darwin-arm64 build-linux build-windows build-all install
