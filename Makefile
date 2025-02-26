.PHONY: release
release:
	goreleaser release --clean
	$(shell echo "$$DOCKER_PASSWORD" | docker login -u akshaykhairmode --password-stdin)
	docker push akshaykhairmode/wscli:$$(git describe --tags --abbrev=0)

.PHONY: test
test:
	goreleaser release --snapshot --clean

.PHONY: lint
lint:
	golangci-lint run ./...