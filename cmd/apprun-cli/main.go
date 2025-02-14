package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	cli "github.com/fujiwara/apprun-cli"
	"golang.org/x/sys/unix"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, unix.SIGTERM)
	defer stop()
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
