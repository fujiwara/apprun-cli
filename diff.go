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

// DiffIgnoreDefault ignores fields that the API never returns.
const DiffIgnoreDefault = ".components[].deploy_source.container_registry.password"

type DiffOption struct {
	Ignore []string `help:"JQ queries to ignore specific fields"`
}

func (c *CLI) runDiff(ctx context.Context) error {
	opt := c.Diff
	local, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		return err
	}
	info, remote, err := c.getApplicationByName(ctx, local.Name)
	if err != nil {
		return err
	}
	id := info.ID
	slog.Info("comparing", "local", c.Application, "remote", id)

	diff, err := diffApplications(id, remote, c.Application, local, opt.Ignore)
	if err != nil {
		return err
	}
	if diff != "" {
		fmt.Print(coloredDiff(diff))
	}
	return nil
}

func diffApplications(remoteName string, remote *Application, localName string, local *Application, extraIgnores []string) (string, error) {
	ignores := []string{DiffIgnoreDefault}
	ignores = append(ignores, extraIgnores...)
	ignore := strings.Join(ignores, ", ") // to be passed to del()
	p, err := gojq.Parse(ignore)
	if err != nil {
		return "", fmt.Errorf("failed to parse ignore query: %s %w", ignore, err)
	}
	diff, err := jsondiff.Diff(
		&jsondiff.Input{Name: remoteName, X: toMap(remote)},
		&jsondiff.Input{Name: localName, X: toMap(local)},
		jsondiff.Ignore(p),
	)
	if err != nil {
		return "", fmt.Errorf("failed to diff: %w", err)
	}
	return diff, nil
}

func coloredDiff(src string) string {
	var b strings.Builder
	for line := range strings.SplitSeq(src, "\n") {
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
