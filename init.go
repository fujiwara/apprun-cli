package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/go-jsonnet/formatter"
)

type InitOption struct {
	Name    string `help:"name of the application to init" required:""`
	Jsonnet bool   `help:"Use jsonnet to generate files"`
}

func (c *CLI) runInit(ctx context.Context) error {
	opt := c.Init
	slog.Info("initializing", "app", opt.Name)
	info, app, err := c.getApplicationByName(ctx, opt.Name)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	slog.Info("found", "id", v(info.Id))
	b, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal application data: %s", err)
	}
	if opt.Jsonnet {
		b, err = jsonToJsonnet(b, "application.jsonnet")
		if err != nil {
			return fmt.Errorf("failed to convert json to jsonnet: %w", err)
		}
		fmt.Print(string(b))
	} else {
		fmt.Println(string(b)) // append newline
	}
	return nil
}

func jsonToJsonnet(src []byte, filepath string) ([]byte, error) {
	s, err := formatter.Format(filepath, string(src), formatter.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to format jsonnet: %w", err)
	}
	return []byte(s), nil
}
