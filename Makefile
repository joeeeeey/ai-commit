
TEST_DIFF_FILE ?= test/diff.txt
.PHONY: build

test:
	export TEST_DIFF_FILE=$(TEST_DIFF_FILE) && \
	go run src/commit_msg_generator.go test/test-commit-msg.txt

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
OUT_FILE := build/commit_msg_generator_$(OS)_$(ARCH)

build:
	go build -o $(OUT_FILE) src/commit_msg_generator.go
	chmod +x $(OUT_FILE)

create_hook:
	chmod +x hook/prepare-commit-msg
	cp hook/prepare-commit-msg .git/hooks/prepare-commit-msg
	cp build/commit_msg_generator_$(OS)_$(ARCH) .git/hooks/commit_msg_generator

cross_platform_build:
	# Build for Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o build/commit_msg_generator_linux_amd64 src/commit_msg_generator.go

	# Build for macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o build/commit_msg_generator_darwin_arm64 src/commit_msg_generator.go
	chmod +x build/commit_msg_generator_darwin_arm64
	chmod +x build/commit_msg_generator_linux_amd64
	# Build for Windows x86_64
	GOOS=windows GOARCH=amd64 go build -o build/commit_msg_generator_win_amd64.exe src/commit_msg_generator.go
