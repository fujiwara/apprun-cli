package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/sacloud/apprun-api-go"
)

type UserOption struct {
	Operation string `arg:"" enum:"create,read" help:"Operation to perform. One of: create, read"`
}

type userAPIMethod func(context.Context) (*http.Response, error)

func (c *CLI) runUser(ctx context.Context) error {
	op := apprun.NewUserOp(c.client)
	var err error
	slog.Info("user operation", "operation", c.User.Operation)
	switch c.User.Operation {
	case "create":
		slog.Info("creating user")
		err = c.callUserAPI(ctx, op.Create)
	case "read":
		slog.Info("reading user")
		err = c.callUserAPI(ctx, op.Read)
	default:
		err = fmt.Errorf("unknown operation: %s", c.User.Operation)
	}
	return err
}

func (c *CLI) callUserAPI(ctx context.Context, fn userAPIMethod) error {
	resp, err := fn(ctx)
	if err != nil {
		return fmt.Errorf("failed to call user API: %s", err)
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	slog.Error("result", "status", status, "status_message", resp.Status, "body", string(body))
	if status >= 400 {
		return fmt.Errorf("failed to call user API with status: %d", status)
	}
	return nil
}
