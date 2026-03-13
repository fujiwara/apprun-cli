package cli

import (
	"context"
	"fmt"
	"log/slog"

	sscli "github.com/fujiwara/sakura-secrets-cli"
	"github.com/fujiwara/tfstate-lookup/tfstate"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func (c *CLI) setupVM(ctx context.Context) error {
	nativeFuncs := DefaultJsonnetNativeFuncs()

	// load secretmanager functions
	secretFunc := sscli.SecretNativeFunction(ctx)
	nativeFuncs = append(nativeFuncs, secretFunc)
	nativeFuncs = append(nativeFuncs, deprecatedSecretValueFunc(secretFunc))

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

// deprecatedSecretValueFunc returns a backward-compatible "secret_value" native function
// that internally delegates to the "secret" function from sakura-secrets-cli.
func deprecatedSecretValueFunc(secretFunc *jsonnet.NativeFunction) *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name: "secret_value",
		Params: []ast.Identifier{
			"vault_id",
			"secret_name",
			"version",
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
			// build name with optional version suffix for the new "secret" function
			name := secretName
			if args[2] != nil {
				v, ok := args[2].(float64)
				if !ok {
					return nil, fmt.Errorf("version must be a number")
				}
				name = fmt.Sprintf("%s:%d", secretName, int(v))
			}
			slog.Warn(
				"secret_value() is deprecated, use secret() instead",
				"migrate_to", fmt.Sprintf(`std.native('secret')('%s', '%s')`, vaultID, name),
			)
			return secretFunc.Func([]any{vaultID, name})
		},
	}
}
