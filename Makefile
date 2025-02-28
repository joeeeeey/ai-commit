
TEST_DIFF_FILE ?= test/diff.txt
.PHONY: test

test:
	export TEST_DIFF_FILE=$(TEST_DIFF_FILE) && \
	go run commit_msg_generator.go test/test-commit-msg.txt