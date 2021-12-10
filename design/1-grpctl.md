# grpctl

## Requirements
- Ability to interact with grpc apis without needing to consult any documentation.
- Ability to customize behaviour for a specific use case if needed.
- Ability to create custom binary for a specific teams grpc services.

## Nice to haves

- No config: grpctl should be able to have no config/paperwork for the user to use the tool.
- Server reflection: can be used like grpcurl, but would support tab completion.

## Inspirations
- [grpcurl](https://github.com/fullstorydev/grpcurl)
  - Universal, doesn't allow to build custom cli.
- [protoc-gen-cobra](https://github.com/fiorix/protoc-gen-cobra)
  - Works for a lot of usecases, but adding code generators have an overhead.
- [gWhisper](https://github.com/IBM/gWhisper)
    - supports tab completion, doesn't support creating a custom cli.

## Ideas

`grpctl` should be a package that allows an engineer to create a custom cli for their teams gRPC API's.
It should use as much information that is possible from the `protodescriptor.FileDescriptorSet` that is possible. This `protoreflect.FileDescriptorSet` can come from the generated go grpc code.
This approach means that there is no need to manually compile `.proto` files to get the `protodescriptor.FileDescriptorSet`.

The conversion of `protodescriptor.FileDescriptorSet` to cobra command would look like this:
- protoreflect.ServiceDescriptor -> top level command (eg `fooctl FooAPI`)
- protoreflect.MethodDescriptor -> second level command (eg `fooctl FooAPI ListBar`)
- protoreflect.MessageDescriptor -> flags (eg `fooctl FooAPI ListBar --field1="string"`)

### Creating a new cli tool

Example of how a new grpctl cli should be created:
```golang
// Example call: billingctl -H="Authorization: Bearer $(gcloud auth application-default print-access-token)" CloudBilling ListBillingAccounts
func main() {
	cmd := &cobra.Command{
		Use:   "billingctl",
		Short: "an example cli tool for the gcp billing api",
	}
	err := grpctl.BuildCommand(cmd,
		grpctl.WithFileDescriptors( // This specifies that we want to 
			billing.File_google_cloud_billing_v1_cloud_billing_proto,
			billing.File_google_cloud_billing_v1_cloud_catalog_proto,
		),
	)
	if err != nil {
		log.Print(err)
	}
	if err := grpctl.RunCommand(cmd, context.Background()); err != nil {
		log.Print(err)
	}
}
```

`grpctl.WithFileDescriptors` allows you to specify the proto descriptors from the generated go code, and in this specific example these FileDescriptors are found [here](https://github.com/googleapis/go-genproto/blob/3a66f561d7aa4010d9715ecf4c19b19e81e19f3c/googleapis/cloud/billing/v1/cloud_billing.pb.go#L767) which are generated from the proto source [here](https://github.com/googleapis/googleapis/blob/987192dfddeb79d3262b9f9f7dbf092827f931ac/google/cloud/billing/v1/cloud_billing.proto).


