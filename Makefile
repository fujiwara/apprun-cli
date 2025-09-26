.PHONY: clean test

build:
	go build -o apprun-cli ./cmd/apprun-cli

clean:
	rm -rf apprun-cli dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/apprun-cli/cmd/apprun-cli

dist:
	goreleaser build --snapshot --clean
