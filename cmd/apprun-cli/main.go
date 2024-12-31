package main

import (
	"context"
	"log/slog"
	"os"

	cli "github.com/fujiwara/apprun-cli"
)

func main() {
	ctx := context.TODO()
	if err := run(ctx); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	c, err := cli.New(ctx)
	if err != nil {
		return err
	}
	return c.Run(ctx)
}
