package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
	progressbar "github.com/schollz/progressbar/v3"
)

type TrafficPercentageByVersion map[string]int

const TrafficShiftDefaultPeriod = time.Minute

type TrafficsOption struct {
	Set               TrafficPercentageByVersion `help:"Set traffic percentage for each version" mapsep:","`
	ShiftTo           string                     `help:"Shift all traffic to the specified version"`
	Rate              int                        `help:"Shift rate percentage(per minute)" default:"100"`
	Period            time.Duration              `help:"Shift period" default:"1m"`
	RollbackOnFailure bool                       `help:"Rollback to the previous version if failed to shift" default:"true" negatable:""`
}

func (c *CLI) runTraffics(ctx context.Context) error {
	opt := c.Traffics
	app, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		return err
	}
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		return err
	}

	if len(opt.Set) > 0 {
		return c.updateTraffics(ctx, v(info.Id), opt.Set)
	}

	if opt.ShiftTo != "" {
		return c.shiftTraffics(ctx, v(info.Id), opt.ShiftTo, opt.Rate, opt.Period)
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

func (c *CLI) shiftTraffics(ctx context.Context, appId string, versionName string, rate int, period time.Duration) error {
	if rate <= 0 || rate > 100 {
		return fmt.Errorf("rate must be between 1 and 100")
	}

	op := apprun.NewTrafficOp(c.client)
	res, err := op.List(ctx, appId)
	if err != nil {
		return err
	}
	if len(*res.Data) > 1 {
		return fmt.Errorf("traffic shifting is not supported for multiple versions")
	}
	data := *res.Data
	currentVersionName := ""
	if v(data[0].IsLatestVersion) {
		slog.Debug("finding latest version")
		vop := apprun.NewVersionOp(c.client)
		param := &v1.ListApplicationVersionsParams{
			SortOrder: ptr(v1.ListApplicationVersionsParamsSortOrder(v1.ListApplicationVersionsParamsSortOrderDesc)),
			PageSize:  ptr(1),
		}
		res, err := vop.List(ctx, appId, param)
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}
		if len(*res.Data) == 0 {
			return fmt.Errorf("no versions found")
		}
		currentVersionName = v((*res.Data)[0].Name)
	} else {
		currentVersionName = v(data[0].VersionName)
	}
	slog.Debug("current traffics", "version", currentVersionName)
	if currentVersionName == versionName {
		slog.Info("already accepts all traffics", "version", versionName)
		return nil
	}

	var completed bool
	defer func() {
		slog.Debug("returning", "completed", completed)
		if !completed && c.Traffics.RollbackOnFailure {
			slog.Info("rolling back traffics", "app", appId, "from", versionName, "to", currentVersionName)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			if err := c.updateTraffics(ctx, appId, TrafficPercentageByVersion{
				currentVersionName: 100,
			}); err != nil {
				slog.Error("failed to rollback traffics", "app", appId, "from", versionName, "to", currentVersionName, "error", err)
				os.Exit(1)
			}
		}
	}()

	slog.Info("shifting traffics", "app", appId, "from", currentVersionName, "to", versionName, "rate", rate, "per", period)

	bar := progressbar.NewOptions(100,
		progressbar.OptionSetDescription("Traffic shifted"),
		progressbar.OptionSetWidth(20),
	)
	shiftedRate := 0
	for {
		shiftedRate += rate
		if shiftedRate >= 100 {
			shiftedRate = 100
		}
		b := v1.PutTrafficsBody{
			{
				VersionName: ptr(versionName),
				Percent:     ptr(shiftedRate),
			},
		}
		if shiftedRate < 100 {
			// Percent == 0 is not allowed...
			b = append(b, v1.Traffic{
				VersionName: ptr(currentVersionName),
				Percent:     ptr(100 - shiftedRate),
			})
		}
		slog.Debug("updating traffics", "traffics", toJSON(b))
		res, err := op.Update(ctx, appId, &b)
		if err != nil {
			return fmt.Errorf("failed to update traffics: %w", err)
		}
		slog.Debug("traffics updated", "traffics", toJSON(res))
		bar.Set(shiftedRate)
		if shiftedRate >= 100 {
			break
		}
		sleep := time.NewTimer(period)
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				slog.Warn("context cancelled", "error", err)
			}
			return nil
		case <-sleep.C:
			// do nothing, next loop
		}
	}
	completed = true
	bar.Finish()
	slog.Info("traffics shifted completely", "app", appId, "from", currentVersionName, "to", versionName)

	return nil
}
