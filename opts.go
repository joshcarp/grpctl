package grpctl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joshcarp/grpctl/ir"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// WithContextFunc will add commands to the cobra command through the file descriptors provided.
func WithFileDescriptors(descriptors ...protoreflect.FileDescriptor) CommandOption {
	return func(cmd *cobra.Command) error {
		err := CommandFromFileDescriptors(cmd, descriptors...)
		if err != nil {
			return err
		}
		return nil
	}
}

// WithContextFunc will modify the context  before the main command is run but not in the completion stage.
func WithContextFunc(f func(context.Context, *cobra.Command) (context.Context, error)) CommandOption {
	return func(cmd *cobra.Command) error {
		existingPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if existingPreRun != nil {
				err := existingPreRun(cmd, args)
				if err != nil {
					return err
				}
			}
			ctx, err := f(cmd.Root().Context(), cmd)
			if err != nil {
				return err
			}
			cmd.Root().SetContext(ctx)
			return nil
		}
		return nil
	}
}

// WithContextDescriptorFunc will modify the context  before the main command is run but not in the completion stage.
func WithContextDescriptorFunc(f func(context.Context, *cobra.Command, protoreflect.MethodDescriptor) (context.Context, error)) CommandOption {
	return func(cmd *cobra.Command) error {
		existingPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if existingPreRun != nil {
				err := existingPreRun(cmd, args)
				if err != nil {
					return err
				}
			}

			a := cmd.Root().Context().Value(methodDescriptorKey{})
			method, ok := a.(protoreflect.MethodDescriptor)
			if !ok {
				return nil
			}
			ctx, err := f(cmd.Root().Context(), cmd, method)
			if err != nil {
				return err
			}
			cmd.Root().SetContext(ctx)
			return nil
		}
		return nil
	}
}

// WithArgs will set the args of the command as args[1:].
func WithArgs(args []string) CommandOption {
	return func(cmd *cobra.Command) error {
		cmd.SetArgs(args[1:])
		return nil
	}
}

// WithReflection will enable grpc reflection on the command. Use this as an alternative to WithFileDescriptors.
func WithReflection(args []string) CommandOption {
	return func(cmd *cobra.Command) error {
		var err error
		cmd.ValidArgsFunction = func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			fds, err2 := reflectFileDesc(args)
			if err2 != nil {
				err = err2
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var opts []string
			err2 = CommandFromFileDescriptors(cmd, fds...)
			if err2 != nil {
				err = err2
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return opts, cobra.ShellCompDirectiveNoFileComp
		}
		fds, err := reflectFileDesc(args[1:])
		if err != nil {
			return err
		}
		if err = persistentFlags(cmd, ""); err != nil {
			return err
		}
		err = CommandFromFileDescriptors(cmd, fds...)
		if err != nil {
			return err
		}
		return nil
	}
}

func WithFern(args []string) CommandOption {
	irep, err := readIR(args[1])
	if err != nil {
		panic(err)
	}
	return func(command *cobra.Command) error {
		err = func(cmd *cobra.Command, irep *ir.IntermediateRepresentation) error {
			for _, service := range irep.Services {
				for _, method := range service.Endpoints {
					if err := CommandFromHttpEndpoint(cmd, method); err != nil {
						return err
					}
				}
			}
			return nil
		}(command, irep)
		if err != nil {
			return err
		}
		print(irep)
		return nil
	}
}

func CommandFromHttpEndpoint(cmd *cobra.Command, endpoint *ir.HttpEndpoint) error {
	cmd.AddCommand(&cobra.Command{
		Use: endpoint.Name,
		Short: endpoint.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	})
	for _, pathParam := range endpoint.AllPathParameters{
		pathParam.
	},
}



// readIR reads the *InermediateRepresentation from the given filename.
func readIR(irFilename string) (*ir.IntermediateRepresentation, error) {
	bytes, err := os.ReadFile(irFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read intermediate representation: %v", err)
	}
	ir := new(ir.IntermediateRepresentation)
	if err := json.Unmarshal(bytes, ir); err != nil {
		return nil, fmt.Errorf("failed to unmarshal intermediate representation: %v", err)
	}
	return ir, nil
}

func WithCompletion() CommandOption {
	return func(command *cobra.Command) error {
		cmd := &cobra.Command{
			Use:   "completion [bash|zsh|fish|powershell]",
			Short: "Generate completion script",
			Long: fmt.Sprintf(`To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s completion bash > /usr/local/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ %[1]s completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

  PS> %[1]s completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %[1]s completion powershell > %[1]s.ps1
  # and source this file from your PowerShell profile.
`, command.Root().Name()),
			DisableFlagsInUseLine: true,
			ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
			Args:                  cobra.ExactValidArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				var err error
				switch args[0] {
				case "bash":
					err = cmd.Root().GenBashCompletion(os.Stdout)
				case "zsh":
					err = cmd.Root().GenZshCompletion(os.Stdout)
				case "fish":
					err = cmd.Root().GenFishCompletion(os.Stdout, true)
				case "powershell":
					err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
				}
				return err
			},
		}
		command.AddCommand(cmd)
		return nil
	}
}


func Get(
	ctx context.Context,
// The ID of the embed job to retrieve.
	id string,
	opts ...option.RequestOption,
) (*v2.EmbedJob, error) {
	options := core.NewRequestOptions(opts...)

	baseURL := "https://api.cohere.ai/v1"
	if c.baseURL != "" {
		baseURL = c.baseURL
	}
	if options.BaseURL != "" {
		baseURL = options.BaseURL
	}
	endpointURL := fmt.Sprintf(baseURL+"/"+"embed-jobs/%v", id)

	headers := core.MergeHeaders(c.header.Clone(), options.ToHeader())

	errorDecoder := func(statusCode int, body io.Reader) error {
		raw, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		apiError := core.NewAPIError(statusCode, errors.New(string(raw)))
		decoder := json.NewDecoder(bytes.NewReader(raw))
		switch statusCode {
		case 400:
			value := new(v2.BadRequestError)
			value.APIError = apiError
			if err := decoder.Decode(value); err != nil {
				return apiError
			}
			return value
		case 404:
			value := new(v2.NotFoundError)
			value.APIError = apiError
			if err := decoder.Decode(value); err != nil {
				return apiError
			}
			return value
		case 429:
			value := new(v2.TooManyRequestsError)
			value.APIError = apiError
			if err := decoder.Decode(value); err != nil {
				return apiError
			}
			return value
		case 500:
			value := new(v2.InternalServerError)
			value.APIError = apiError
			if err := decoder.Decode(value); err != nil {
				return apiError
			}
			return value
		}
		return apiError
	}

	var response *v2.EmbedJob
	if err := c.caller.Call(
		ctx,
		&core.CallParams{
			URL:          endpointURL,
			Method:       http.MethodGet,
			MaxAttempts:  options.MaxAttempts,
			Headers:      headers,
			Client:       options.HTTPClient,
			Response:     &response,
			ErrorDecoder: errorDecoder,
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}