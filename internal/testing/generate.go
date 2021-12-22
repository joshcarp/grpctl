package testing

//go:generate protoc -I proto --go_out=paths=source_relative:proto/examplepb --go-grpc_out=paths=source_relative:proto/examplepb proto/api.proto
