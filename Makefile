.phony: build
build:
	rm -rf bin && mkdir bin
	@echo "Building..."
	GOOS=windows GOARCH=386 go build -o bin/windows-intel/wscli.exe main.go
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/wscli.exe main.go

	GOOS=linux GOARCH=amd64 go build -o bin/linux/wscli main.go
	GOOS=linux GOARCH=386 go build -o bin/linux-intel/wscli main.go

	GOOS=darwin GOARCH=amd64 go build -o bin/darwin-intel/wscli main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-apple-silicone/wscli main.go

	zip bin/binaries.zip \
	bin/windows-intel/wscli.exe \
	bin/windows-amd64/wscli.exe \
	bin/linux/wscli \
	bin/linux-intel/wscli \
	bin/darwin-intel/wscli \
	bin/darwin-apple-silicone/wscli