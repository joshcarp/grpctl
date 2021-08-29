package grpctl

import (
	"fmt"
	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

type Config struct {
	ConfigFile     string       `yaml:"-"`
	CurrentContext string       `yaml:"current-context"`
	Contexts       []RawContext `yaml:"contexts"`
	Users          []User       `yaml:"users"`
	Services       []Services   `yaml:"services"`
}

func (c Config) GetCurrentContext() Context {
	return GetCurrentContext(c)
}

func (c Config) GetServiceConfig(service string) (string, bool, bool) {
	for _, e := range c.Services{
		if e.Name == service {
			return e.Addr, e.Plaintext, true
		}
	}
	return "", false, false
}

func (c Config) GetMethodConfig(service, method string) []byte {
	for _, e := range c.Services {
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

func (c Config) Save() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(c.ConfigFile, b, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func LoadConfig(filename string) (Config, error) {
	config := Config{ConfigFile: filename}
	b, err := os.ReadFile(filename)
	if err != nil {
		cobra.CheckErr(err)
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
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
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Methods struct {
	Name       string `yaml:"name"`
	Descriptor []byte `yaml:"descriptor"`
}

type Services struct {
	Name         string            `yaml:"name"`
	Environments map[string]string `yaml:"environments"`
	Addr         string            `yaml:"addr"`
	Plaintext    bool              `yaml:"plaintext"`
	Descriptor   []byte            `yaml:"descriptor"`
	Methods      []Methods         `yaml:"methods"`
}

func NewService(service protoreflect.ServiceDescriptor){

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
			return curCtx
		}
	}
	return Context{}
}

func initConfig(cfgFile string) Config {
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}
		cfgFile = path.Join(home, ".grpctl.yaml")
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			err := Config{ConfigFile: cfgFile}.Save()
			cobra.CheckErr(err)
		}
	}
	config, err := LoadConfig(cfgFile)
	cobra.CheckErr(err)
	return config
}
