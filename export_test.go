package cli

import "context"

func (c *CLI) SetupVM(ctx context.Context) error {
	return c.setupVM(ctx)
}

var FromV1Application = fromV1Application

var DiffApplications = diffApplications
