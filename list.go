package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

type ListOption struct {
}

func (c *CLI) runList(ctx context.Context) error {
	for data, err := range c.allApplications(ctx) {
		if err != nil {
			return fmt.Errorf("failed to list applications: %s", err)
		}
		if b, err := json.MarshalIndent(data, "", "  "); err != nil {
			return fmt.Errorf("failed to marshal application data: %s", err)
		} else {
			fmt.Println(string(b))
		}
	}
	return nil
}

var ErrNotFound = fmt.Errorf("not found")

func (c *CLI) allApplications(ctx context.Context) func(func(*v1.HandlerListApplicationsData, error) bool) {
	op := apprun.NewApplicationOp(c.client)
	param := &v1.ListApplicationsParams{
		SortField: ptr("name"),
		SortOrder: ptr(v1.ListApplicationsParamsSortOrder(v1.HandlerListApplicationsMetaSortOrderAsc)),
		PageSize:  ptr(100),
	}
	var page int
	return func(yield func(*v1.HandlerListApplicationsData, error) bool) {
		for {
			page++
			param.PageNum = ptr(page)
			slog.Debug("fetching list applications", "page", page)
			res, err := op.List(ctx, param)
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
