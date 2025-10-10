package cli

import (
	"context"
	"fmt"

	"github.com/fujiwara/tfstate-lookup/tfstate"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

func (c *CLI) setupVM(ctx context.Context) error {
	nativeFuncs := DefaultJsonnetNativeFuncs()
	// load secretmanager functions
	funcs, err := secretsManagerNativeFuncs(ctx)
	if err != nil {
		return err
	}
	nativeFuncs = append(nativeFuncs, funcs...)

	// load tfstate functions
	if c.TFState != "" {
		lookup, err := tfstate.ReadURL(ctx, c.TFState)
		if err != nil {
			return err
		}
		nativeFuncs = append(nativeFuncs, lookup.JsonnetNativeFuncs(ctx)...)
	}

	c.loader.AddFunctions(nativeFuncs...)
	return nil
}

func secretsManagerNativeFuncs(ctx context.Context) ([]*jsonnet.NativeFunction, error) {
	client, err := sm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	return []*jsonnet.NativeFunction{
		{
			Name: "secret_value",
			Params: []ast.Identifier{
				"vault_id",    // vault_id is a resource ID of SecretManager Vault
				"secret_name", // secret_name is a name of secret
				"version",     // version is a version number of secret, null means latest version
			},
			Func: func(args []any) (any, error) {
				vaultID, ok := args[0].(string)
				if !ok {
					return nil, fmt.Errorf("vault_id must be a string")
				}
				secretName, ok := args[1].(string)
				if !ok {
					return nil, fmt.Errorf("secret_name must be a string")
				}
				var version v1.OptNilInt
				if args[2] != nil {
					v, ok := args[2].(float64)
					if !ok {
						return nil, fmt.Errorf("version must be a number")
					}
					version = v1.NewOptNilInt(int(v))
				}
				secOp := sm.NewSecretOp(client, vaultID)
				res, err := secOp.Unveil(ctx, v1.Unveil{
					Name:    secretName,
					Version: version,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to unveil secret with vault_id: %s secret_name: %s version: %v: %s", vaultID, secretName, args[2], err)
				}
				return res.Value, nil
			},
		},
	}, nil
}
