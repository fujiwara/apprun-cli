package cli

import (
	"context"
	"fmt"
	"log/slog"

	apprun "github.com/sacloud/apprun-api-go"
)

type UserOption struct {
	Operation string `arg:"" enum:"create,read" help:"Operation to perform. One of: create, read"`
}

func (c *CLI) runUser(ctx context.Context) error {
	op := apprun.NewUserOp(c.client)
	slog.Info("user operation", "operation", c.User.Operation)
	switch c.User.Operation {
	case "create":
		slog.Info("creating user")
		res, err := op.Create(ctx)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		fmt.Println(toJSONIndent(res))
		return nil
	case "read":
		slog.Info("reading user")
		res, err := op.Read(ctx)
		if err != nil {
			return fmt.Errorf("failed to read user: %w", err)
		}
		fmt.Println(toJSONIndent(res))
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", c.User.Operation)
	}
}
