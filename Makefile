.PHONY: clean test

clean:
	rm -rf apprun-cli dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/apprun-cli/cmd/apprun-cli

dist:
	goreleaser build --snapshot --clean
