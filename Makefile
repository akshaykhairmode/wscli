.PHONY: install
install:
	go install github.com/goreleaser/goreleaser/v2@latest

.PHONY: release
release:
	goreleaser release --clean
	bash docker.sh

.PHONY: test
test:
	goreleaser release --snapshot --clean

.PHONY: lint
lint:
	golangci-lint run ./...