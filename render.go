package cli

import (
	"context"
	"encoding/json"
	"fmt"
)

type RenderOption struct {
	Jsonnet bool `help:"Format as Jsonnet to render files"`
}

func (c *CLI) runRender(ctx context.Context) error {
	opt := c.Render
	app, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		return fmt.Errorf("failed to load application: %w", err)
	}
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
