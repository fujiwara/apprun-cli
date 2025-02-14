package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

type TrafficPercentageByVersion map[string]int

type TrafficsOption struct {
	Application string                     `arg:"" name:"application" help:"Name of the definition file to use"`
	Versions    TrafficPercentageByVersion `help:"Traffic percentage for each version" mapsep:","`
}

func (c *CLI) runTraffics(ctx context.Context) error {
	opt := c.Traffics
	app, err := LoadApplication(ctx, opt.Application)
	if err != nil {
		return err
	}
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		return err
	}

	if len(opt.Versions) > 0 {
		return c.updateTraffics(ctx, v(info.Id), opt.Versions)
	}

	for tr, err := range c.AllTraffics(ctx, v(info.Id)) {
		if err != nil {
			return err
		}
		fmt.Println(toJSONIndent(tr))
	}
	return nil
}

func (c *CLI) AllTraffics(ctx context.Context, appId string) func(func(*v1.Traffic, error) bool) {
	op := apprun.NewTrafficOp(c.client)
	return func(yield func(*v1.Traffic, error) bool) {
		for {
			res, err := op.List(ctx, appId)
			if err != nil {
				yield(nil, err)
				return
			}
			if len(*res.Data) == 0 {
				return
			}
			for _, data := range *res.Data {
				if !yield(&data, nil) {
					return
				}
			}
			return
		}
	}
}

func (c *CLI) updateTraffics(ctx context.Context, appId string, versions TrafficPercentageByVersion) error {
	slog.Info("updating traffics", "app", appId, "traffics", toJSON(versions))
	op := apprun.NewTrafficOp(c.client)
	b := v1.PutTrafficsBody{}
	for version, percentage := range versions {
		b = append(b, v1.Traffic{
			VersionName: ptr(version),
			Percent:     ptr(percentage),
		})
	}
	res, err := op.Update(ctx, appId, &b)
	if err != nil {
		return fmt.Errorf("failed to update traffics: %w", err)
	}
	fmt.Println(toJSONIndent(res))
	return nil
}
