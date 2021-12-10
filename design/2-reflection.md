# grpctl support for reflection

## Requirements
- Ability to interact with reflection apis.
- No config, and no paging input. This means that any command written can be shared without any external configuration.
- Support for tab completion

## Ideas

reflection is just another way of getting hold of the `protodescriptor.FileDescriptor` as opposed to explicitly importing the `pb.File_foo` from the generated go code.

In order to meet the requirement of no config with tab completion, the cli tool would need to accomplish gRPC server reflection within the tab completion stage.
For example once the following is written, and then `tab` is pressed:
```bash
grpctl --address=localhost:8081 --plaintext=true [tab-tab]
```

grpctl would go out to the grpc server located at `localhost:8081` and hit its server reflection api.

After this initial part, the `protodescriptor.FileDescriptor` would be cached locally so the next time tab completion is needed it does not need to use grpc reflection.

After this stage the cli is identical to using the static `protodescriptor.FileDescriptor`.

