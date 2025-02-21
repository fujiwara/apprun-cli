package cli

import (
	"context"

	"github.com/google/go-jsonnet"
	"github.com/sacloud/apprun-api-go"
)

func New(ctx context.Context) (*CLI, error) {
	vm := jsonnet.MakeVM()
	nativeFuncs := DefaultJsonnetNativeFuncs()
	for _, f := range nativeFuncs {
		vm.NativeFunction(f)
	}
	return &CLI{
		client: &apprun.Client{},
		vm:     vm,
	}, nil
}
