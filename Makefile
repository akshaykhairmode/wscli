.phony: build
build:
	@echo "Building..."
	GOOS=windows GOARCH=386 go build -o bin/windows-intel/wscli.exe main.go interactive.go
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/wscli.exe main.go interactive.go

	GOOS=linux GOARCH=amd64 go build -o bin/linux/wscli main.go interactive.go
	GOOS=linux GOARCH=386 go build -o bin/linux-intel/wscli main.go interactive.go

	GOOS=darwin GOARCH=amd64 go build -o bin/darwin-intel/wscli main.go interactive.go
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-apple-silicone/wscli main.go interactive.go