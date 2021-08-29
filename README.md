# grpctl

A dynamic cli for interacting with grpc apis. Sort of like a mash of [grpcurl](https://github.com/fullstorydev/grpcurl) and [kubectl](https://github.com/kubernetes/kubectl).
This project was inspired by [protoc-gen-cobra](https://github.com/fiorix/protoc-gen-cobra) but sometimes adding another protoc plugin is annoying.

## How does it work?
Instead of manually writing or code generating cobra commands grpctl uses the `protoreflect.FileDescriptor` to interact with services, methods and types. 

The mapping is something like this:
- protoreflect.ServiceDescriptor -> top level command (eg `fooctl FooAPI`)
- protoreflect.MethodDescriptor -> second level command (eg `fooctl FooAPI ListBar`)
- protoreflect.MessageDescriptor -> flags (eg `fooctl FooAPI ListBar --field1="string"`)

This also means that autocomplete example payloads can be generated.

## Reflection mode
This mode is for using grpctl with reflection apis.

### Install
```bash
go get github.com/joscharp/grpctl/cmd/grpctl
```

### Run
```bash
> grpctl add --addr localhost:8080 --plaintext
> grpctl list
Health Alive
Health Ready
Health Version
FooAPI ListBar
BarAPI GetFoo

> grpctl --help
> grpctl
FooAPI            -- FooAPI as defined in foo/bar.proto
BarAPI            -- BarAPI as defined in bar/foo.proto
add               -- Add a services to grpctl
completion        -- generate the autocompletion script for the specified shell
config            -- configure options in grpctl
help              -- Help about any command
list              -- list service
remove            -- remove a service from saved services
```


## File descriptor mode
This mode is for creating an api specific cli tool (like kubectl).

### Install

```go
package main

import (
	"github.com/joshcarp/grpcexample/proto/examplepb"
	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "examplectl",
		Short: "a cli tool for example",
	}
	grpctl.Execute(cmd, examplepb.File_api_proto)
}

```

this will use the service and method descriptors in `altethical.File_api_proto` to dynamically create cobra commands:

```bash
> altethicalctl --help           
a cli tool for altethical

Usage:
  altethical [command]

Available Commands:
  altethical  altethical as defined in api.proto
  completion  generate the autocompletion script for the specified shell
  config      configure options in grpctl
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.grpctl.yaml)
  -h, --help            help for altethical
  -t, --toggle          Help message for toggle

Use "altethical [command] --help" for more information about a command.

> altethicalctl altethical --help
altethical as defined in api.proto

Usage:
  altethical altethical [command]

Available Commands:
  byclicks    byclicks as defined in api.proto
  byimages    byimages as defined in api.proto
  example     example as defined in api.proto
  searchImage searchImage as defined in api.proto

Flags:
  -h, --help   help for altethical

Global Flags:
      --config string   config file (default is $HOME/.grpctl.yaml)

Use "altethical altethical [command] --help" for more information about a command.

> altethicalctl altethical example --message foo --addr foobar.com
Hello foo!  
```

## Features
- [x] Dynamic generation cobra commands for grpc Services and `Methods`.
- [x] Generation of flags for top level input types.
- [x] Generation of auto completion for types.
- [x] Proto file descriptor support. 
- [x] gRPC reflection support.