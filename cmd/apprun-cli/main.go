package main

import (
	"context"
	"log"

	cli "github.com/fujiwara/apprun-cli"
)

func main() {
	ctx := context.TODO()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	return cli.Run(ctx)
}
