package grpctl

import (
	"context"

	"github.com/spf13/cobra"
)

type customContext struct {
	ctx *context.Context
}

type (
	configKeyType           struct{}
	methodDescriptorKeyType struct{}
)

var (
	configKey           = &configKeyType{}
	methodDescriptorKey = &methodDescriptorKeyType{}
)

func (c *customContext) setContext(ctx context.Context) {
	*(c).ctx = ctx
}

func getContext(cmd *cobra.Command) (*customContext, context.Context, bool) {
	ctx := cmd.Root().Context()
	val := ctx.Value(configKey)
	ctx2, ok := val.(*customContext)
	if !ok {
		return nil, nil, ok
	}
	return ctx2, *ctx2.ctx, ok
}
