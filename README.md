<h1 align="center">grpctl</h1>

<div align="center">

[![Status](https://img.shields.io/badge/status-active-success.svg)]()
[![GitHub Issues](https://img.shields.io/github/issues/joshcarp/grpctl)](https://github.com/joshcarp/grpctl/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/joshcarp/grpctl)](https://github.com/joshcarp/grpctl/pulls)
[![License](https://img.shields.io/badge/license-apache2-blue.svg)](/LICENSE)

</div>

A golang package for easily creating custom cli tools from FileDescriptors, or through the gRPC reflection API. 

# 📖 Table of contents

- [Reflection cli mode](#reflection-cli-mode)
- [File descriptor mode](#file-descriptor-mode)
- [Autocompletion](#autocompletion)
- [Flags](#flags)
- [Design](#design)
- [Contributing](#contributing)
- [License](#license)

## 🪞 Reflection cli mode <a name = "reflection-cli-mode"></a>

To be used like `grpcurl` against reflection APIs but with tab completion.

![grpctl](./grpctl.svg)

### 📥 Install

```bash
go get github.com/joshcarp/grpctl/cmd/grpctl
grpctl --help
```

[embedmd]:# (cmd/grpctl/docs/grpctl.md bash /  -a/ /WithInsecure/)
```bash
  -a, --address string       Address in form 'host:port'
      --config string        Config file (default is $HOME/.grpctl.yaml)
  -H, --header stringArray   Header in form 'key: value'
  -h, --help                 help for grpctl
  -p, --plaintext            Dial grpc.WithInsecure
```

## 🗄️ File descriptor mode <a name = "file-descriptor-mode"></a>

To easily create a cli tool for your grpc APIs using the code generated `protoreflect.FileDescriptor`
To view all options that can be used, see [opts.go](opts.go).

![examplectl](./examplectl.gif)

### 📥 Install

[embedmd]:# (cmd/billingctl/main.go go /func main/ $)
```go
func main() {
	cmd := &cobra.Command{
		Use:   "billingctl",
		Short: "an example cli tool for the gcp billing api",
	}
	err := grpctl.BuildCommand(cmd,
		grpctl.WithArgs(os.Args),
		grpctl.WithFileDescriptors(
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

## 🤖 Autocompletion <a name = "autocompletion"></a>

run `grpctl completion --help` and do what it says

## 🏳️‍🌈 Flags <a name = "flags"></a>

- `--address`
```bash
grpctl --address=<scheme://host:port>
```
  - it is important that the `=` is used with flags, otherwise the value will be interpreted as a command which does not exist.

- `--header`
```bash
grpctl --address=<scheme://host:port> -H="Foo:Bar" -H="Bar: Foo"
```
  - Any white spaces at the start of the value will be stripped

- `--protocol`
```bash
grpctl --address=<scheme://host:port> --protocol=<connect|grpc|grpcweb>
```
- Specifies which rpc protocol to use, default=grpc

- `--http1`
```bash
grpctl --address=<scheme://host:port> --http1
```
- Use a http1.1 client instead of http2

# 🧠 Design <a name = "design"></a>

Design documents (more like a stream of consciousness) can be found in [./design](./design).

# 🔧 Contributing <a name = "contributing"></a>

This project is still in an alpha state, any contributions are welcome see [CONTRIBUTING.md](CONTRIBUTING.md).

There is also a slack channel on gophers slack: [#grpctl](https://gophers.slack.com/archives/C02CAH9NP7H)

# 🖋️ License <a name = "license"></a>

See [LICENSE](LICENSE) for more details.

## 🎉 Acknowledgements <a name = "acknowledgement"></a>
- [@dancantos](https://github.com/dancantos)/[@anzboi](https://github.com/anzboi) and I were talking about [protoc-gen-cobra](https://github.com/fiorix/protoc-gen-cobra) when dan came up with the idea of using the proto descriptors to generate cobra commands on the fly.
 
