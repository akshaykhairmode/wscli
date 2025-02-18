.phony: release
release:
	goreleaser release --clean

.phony: test
test:
	goreleaser release --snapshot --clean