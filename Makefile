
TEST_DIFF_FILE ?= test/diff.txt
.PHONY: test

test:
	export TEST_DIFF_FILE=$(TEST_DIFF_FILE) && \
	go run commit_msg_generator.go test/test-commit-msg.txt

build:
	go build -o bin/commit_msg_generator .
	chmod +x bin/prepare-commit-msg
	chmod +x bin/commit_msg_generator

create_hook:
	cp bin/prepare-commit-msg .git/hooks/prepare-commit-msg
	cp bin/commit_msg_generator .git/hooks/commit_msg_generator
