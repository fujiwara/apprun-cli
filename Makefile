.PHONY: clean test

build:
	go build -tags no_gcs,no_azurerm -o apprun-cli ./cmd/apprun-cli

clean:
	rm -rf apprun-cli dist/

test:
	go test -v ./...

install:
	go install -tags no_gcs,no_azurerm github.com/fujiwara/apprun-cli/cmd/apprun-cli

dist:
	goreleaser build --snapshot --clean
