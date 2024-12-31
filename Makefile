.PHONY: clean test

apprun-cli: go.* *.go
	go build -o $@ cmd/apprun-cli/main.go

clean:
	rm -rf apprun-cli dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/apprun-cli/cmd/apprun-cli

dist:
	goreleaser build --snapshot --clean
