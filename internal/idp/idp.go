package idp

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/pkg/cmd/create"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/delete"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
)

// SetLogger configures the in-process idpbuilder logger. It must be called
// before Run or Delete; idpbuilder's create/delete PreRunE re-applies it.
func SetLogger(level string, colored bool) error {
	helpers.LogLevel = level
	helpers.ColoredOutput = colored
	return helpers.SetLogger()
}

// Run executes idpbuilder's create engine in-process. args are the idpbuilder
// create flags — the same surface as the curated flags and `--` passthrough.
func Run(ctx context.Context, args []string) error {
	cmd := create.CreateCmd
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("idpbuilder create failed: %w", err)
	}
	return nil
}

// Delete tears down the IDP Kind cluster by name (defaults to localdev).
func Delete(ctx context.Context, name string) error {
	cmd := delete.DeleteCmd
	if name != "" {
		cmd.SetArgs([]string{"--name", name})
	} else {
		cmd.SetArgs(nil)
	}
	cmd.SetContext(ctx)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("idpbuilder delete failed: %w", err)
	}
	return nil
}
