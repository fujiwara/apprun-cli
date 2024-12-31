package cli

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aereal/jsondiff"
	"github.com/fatih/color"
	"github.com/itchyny/gojq"
)

const DiffIgnoreDefault = ".components[].deploy_source.container_registry.password"

type DiffOption struct {
	Application string   `arg:"" help:"Name of the definition file to diff" required:""`
	Ignore      []string `help:"JQ queries to ignore specific fields"`
}

func (c *CLI) runDiff(ctx context.Context) error {
	opt := c.Diff
	local, err := LoadApplication(ctx, opt.Application)
	if err != nil {
		return err
	}
	info, remote, err := c.getApplicationByName(ctx, local.Name)
	if err != nil {
		return err
	}
	id := v(info.Id)
	slog.Info("comparing", "local", opt.Application, "remote", id)

	opts := []jsondiff.Option{}
	ignores := []string{}
	ignores = append(ignores, DiffIgnoreDefault)
	ignores = append(ignores, opt.Ignore...)
	ignore := strings.Join(ignores, " or ")
	if p, err := gojq.Parse(ignore); err != nil {
		return fmt.Errorf("failed to parse ignore query: %s %w", ignore, err)
	} else {
		opts = append(opts, jsondiff.Ignore(p))
	}

	if diff, err := jsondiff.Diff(
		&jsondiff.Input{Name: id, X: toMap(remote)},
		&jsondiff.Input{Name: opt.Application, X: toMap(local)},
		opts...,
	); err != nil {
		return fmt.Errorf("failed to diff: %w", err)
	} else if diff != "" {
		fmt.Print(coloredDiff(diff))
	}
	return nil
}

func coloredDiff(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, "-") {
			b.WriteString(color.RedString(line) + "\n")
		} else if strings.HasPrefix(line, "+") {
			b.WriteString(color.GreenString(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}
