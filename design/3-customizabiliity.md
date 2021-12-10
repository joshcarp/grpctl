# customizability in grpctl

## Requirements

- Ability to customize grpctl and add small units of functionality based on the use case.

## Nice to haves

- Should work for a substantial project; something like googleapis.
- Avoid all global state.

## Inspirations

- [functional options](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)

## Ideas

functional options can be used to reduce the complexity of the discrete units of logic. The `cobra.Command` already has
a lot of functionality that can be achieved through `cobra.Command.PersistentPreRun` and `cobra.Command.PreRun`.

One issue that does exist is that the cobra.Command does not have the ability to set the `cobra.Context()`, which means
that customizing behaviour would be limited.

### workaround for lack of setting ability for cobra.Command.Context()

Instead of using the context directly a "custom context" struct can be added
in `cobra.Command.ExecuteContext(context.Contetx)` with a pointer to a mutable context.

Execution would look like this:

```go
func RunCommand(cmd *cobra.Command, ctx context.Context) error {
customCtx := customContext{
ctx: &ctx,
}
return cmd.ExecuteContext(context.WithValue(context.Background(), configKey, &customCtx))
}
```

because `customCtx` stores a pointer to `ctx` one can modify what is in that position, essentially allowing for a
context to be modified as if `cobra.Command.SetContext(context.Context)` existed.

This allows the following to be possible:

```go
&cobra.Command{
    Use: "foobar",
    PreRunE: func(cmd *cobra.Command, args []string) error {
        custCtx, ctx, ok := getContext(cmd) // unwrap the "custom context"
        if !ok {
            return nil
        }
        custCtx.setContext(context.WithValue(ctx, "foo", "bar")) // replace the context that the cusomContext.ctx points to
        return nil
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        _, ctx, ok := getContext(cmd)
        require.True(t, ok, "custom context not found")
        require.Equal(t, "bar", ctx.Value("foo")) // The value is reflected here
        return nil
    },
}

```

## Functional options that edit the context

Now that the context can be edited from prerun hooks Authentication becomes easier, and will not halt autocompletion because PreRun hooks are not executed in the completion stage.

Now  the following functional option is possible:
```go
WithContextFunc(func(ctx context.Context, cmd *cobra.Command) (context.Context, error) {
    return metadata.AppendToOutgoingContext(ctx, "fookey", "fooval"), nil
})
```
This can be used for setting grpc and authentication headers.