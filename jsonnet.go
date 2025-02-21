package cli

import (
	"context"

	"github.com/fujiwara/tfstate-lookup/tfstate"
)

func (c *CLI) setupVM(ctx context.Context) error {
	nativeFuncs := DefaultJsonnetNativeFuncs()

	// load tfstate functions
	if c.TFState != "" {
		lookup, err := tfstate.ReadURL(ctx, c.TFState)
		if err != nil {
			return err
		}
		nativeFuncs = append(nativeFuncs, lookup.JsonnetNativeFuncs(ctx)...)
	}

	for _, f := range nativeFuncs {
		c.vm.NativeFunction(f)
	}
	return nil
}
