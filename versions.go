package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Songmu/prompter"
	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

type VersionsOption struct {
	ID          string `help:"Show the detailed information of the specified version id"`
	Delete      bool   `help:"Delete the specified version id"`
	Force       bool   `help:"Force delete without confirmation"`
}

func (c *CLI) runVersions(ctx context.Context) error {
	opt := c.Versions
	app, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		return err
	}
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		return err
	}

	if opt.ID != "" {
		return c.runVersionID(ctx, opt, info)
	}

	for version, err := range c.AllVersions(ctx, v(info.Id)) {
		if err != nil {
			return err
		}
		fmt.Println(toJSONIndent(version))
	}
	return nil
}

func (c *CLI) runVersionID(ctx context.Context, opt VersionsOption, info *ApplicationInfo) error {
	op := apprun.NewVersionOp(c.client)
	if opt.Delete {
		if !opt.Force && !prompter.YN(fmt.Sprintf("Do you really want to delete version %s of application %s?", opt.ID, *info.Name), false) {
			slog.Info("canceled")
			return nil
		}
		slog.Info("deleting version", "id", opt.ID, "app", *info.Name)
		err := op.Delete(ctx, v(info.Id), opt.ID)
		if err != nil {
			return fmt.Errorf("failed to delete version: %w", err)
		}
		slog.Info("deleted version", "id", opt.ID, "app", *info.Name)
		return nil
	}
	res, err := op.Read(ctx, v(info.Id), opt.ID)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	fmt.Println(toJSONIndent(res))
	return nil
}

func (c *CLI) AllVersions(ctx context.Context, appId string) func(func(*v1.Version, error) bool) {
	op := apprun.NewVersionOp(c.client)
	param := &v1.ListApplicationVersionsParams{
		SortOrder: ptr(v1.ListApplicationVersionsParamsSortOrder(v1.ListApplicationVersionsParamsSortOrderAsc)),
		PageSize:  ptr(100),
	}
	var page int
	return func(yield func(*v1.Version, error) bool) {
		for {
			page++
			param.PageNum = ptr(page)
			slog.Debug("fetching list versions", "app_id", appId, "page", page)
			res, err := op.List(ctx, appId, param)
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
		}
	}
}
