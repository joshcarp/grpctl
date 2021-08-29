package grpctl

import (
	"github.com/joshcarp/grpctl/internal/descriptors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Config struct {
	ConfigFile     string        `yaml:"-"`
	CurrentContext string        `yaml:"current-context"`
	Contexts       []RawContext  `yaml:"contexts"`
	Users          []User        `yaml:"users"`
	Environments   []Environment `yaml:"environments"`
}

func (c Config) GetCurrentContext() Context {
	return GetCurrentContext(c)
}

func (c Config) GetServiceConfig(service string) (string, bool, bool) {
	for _, e := range c.GetCurrentContext().Environment.Services {
		if e.Name == service {
			return e.Addr, e.Plaintext, true
		}
	}
	return "", false, false
}

func (c Config) GetMethodConfig(service, method string) []byte {
	for _, e := range c.GetCurrentContext().Environment.Services {
		if e.Name == service {
			for _, ee := range e.Methods {
				if ee.Name == method {
					return ee.Descriptor
				}
			}
		}
	}
	return nil
}

type User struct {
	Name    string   `yaml:"name"`
	Headers []Header `yaml:"headers"`
}

type RawContext struct {
	Name        string `yaml:"name"`
	User        string `yaml:"user"`
	Environment string `yaml:"environment"`
}

type Context struct {
	Name        string      `yaml:"name"`
	User        User        `yaml:"user"`
	Environment Environment `yaml:"environment"`
}

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Methods struct {
	Name       string `yaml:"name"`
	Descriptor []byte `yaml:"descriptor"`
}

type Environment struct {
	Name     string     `yaml:"name"`
	Services []Services `yaml:"services"`
}

type Services struct {
	Name       string    `yaml:"name"`
	Addr       string    `yaml:"addr"`
	Plaintext  bool      `yaml:"plaintext"`
	Descriptor []byte    `yaml:"descriptor"`
	Methods    []Methods `yaml:"methods"`
}

func (s Services) ServiceDescriptor() (protoreflect.ServiceDescriptor, error) {
	spb := &descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(s.Descriptor, spb)
	if err != nil {
		return nil, err
	}
	desc, err := ConvertToProtoReflectDesc(spb)
	if err != nil {
		return nil, err
	}
	for _, e := range desc {
		for _, ee := range descriptors.NewFileDescriptor(e).Services() {
			if s.Name == ee.Command() {
				return ee.ServiceDescriptor, nil
			}
		}
	}
	return nil, nil
}

func GetCurrentContext(cfg Config) Context {
	for _, e := range cfg.Contexts {
		if e.Name == cfg.CurrentContext {
			var curCtx Context
			for _, ee := range cfg.Users {
				if e.User == ee.Name {
					curCtx.User = ee
				}
			}
			for _, env := range cfg.Environments {
				curCtx.Environment = env
			}
			return curCtx
		}
	}
	return Context{}
}
