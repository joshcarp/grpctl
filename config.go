package grpctl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigFile     string    `json:"-" yaml:"-"`
	CurrentContext string    `json:"current-context" yaml:"current-context"`
	Contexts       []Context `json:"contexts" yaml:"contexts"`
	Users          []User    `json:"users" yaml:"users"`
	Services       []Service `json:"services" yaml:"services"`
}

type User struct {
	Name    string            `json:"name" yaml:"name"`
	Headers map[string]string `json:"headers" yaml:"headers"`
}

type Context struct {
	Name            string `json:"name" yaml:"name"`
	UserName        string `json:"user" yaml:"user"`
	EnvironmentName string `json:"env" yaml:"env"`
	User            User   `json:"-" yaml:"-"`
}

type Header struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

type Methods struct {
	Name string `json:"name" yaml:"name"`
}

type Environment struct {
	Name      string `json:"name" yaml:"name"`
	Addr      string `json:"addr" yaml:"addr"`
	Plaintext bool   `json:"plaintext" yaml:"plaintext"`
}

type Service struct {
	Parent       *Config       `json:"-" yaml:"-"`
	Environments []Environment `json:"environments" yaml:"environments"`
	Name         string        `json:"name" yaml:"name"`
	Descriptor   string        `json:"descriptor" yaml:"descriptor"`
	Methods      []Methods     `json:"methods" yaml:"methods"`
}

func (s Service) String() string {
	s.Descriptor = ""
	marshal, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func (c Config) GetCurrentContext(name string) (Context, error) {
	ctx, err := c.GetContext(name)
	if err != nil {
		return Context{}, err
	}
	ctx.User, err = c.GetUser(ctx.UserName)
	if err != nil {
		return Context{}, err
	}
	return ctx, nil
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

func (c Service) Save() error {
	cfg, err := c.Parent.UpdateService(c)
	if err != nil {
		return err
	}
	return cfg.Save()
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
	for i := range config.Services {
		config.Services[i].Parent = &config
	}
	return config, nil
}

func NewService(fd *descriptorpb.FileDescriptorSet, service protoreflect.ServiceDescriptor, addr string, plaintext bool) Service {
	fdbytes, err := proto.Marshal(fd)
	cobra.CheckErr(err)
	var methods []Methods
	for _, e := range descriptors.NewServiceDescriptor(service).Methods() {
		methods = append(methods, Methods{Name: e.Command()})
	}
	return Service{
		Name: descriptors.NewServiceDescriptor(service).Command(),
		Environments: []Environment{
			{
				Name:      "default",
				Addr:      addr,
				Plaintext: plaintext,
			},
		},
		Descriptor: base64.StdEncoding.EncodeToString(fdbytes),
		Methods:    methods,
	}
}

func (s Service) ServiceDescriptor() (protoreflect.ServiceDescriptor, error) {
	spb := &descriptorpb.FileDescriptorSet{}
	descbytes, err := base64.StdEncoding.DecodeString(s.Descriptor)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(descbytes, spb)
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

func DefaultContext() Context {
	return Context{
		Name:            "default",
		UserName:        "user",
		EnvironmentName: "env",
	}
}

func DefaultUser() User {
	return User{
		Name:    "name",
		Headers: map[string]string{"key": "value"},
	}
}

func DefaultService() Service {
	return Service{
		Environments: []Environment{{
			Name:      "name",
			Addr:      "addr",
			Plaintext: false,
		}},
		Name:       "name",
		Descriptor: "",
		Methods: []Methods{{
			Name: "name",
		}},
	}
}

func DefaultEnvironment() Environment {
	return Environment{
		Name:      "name",
		Addr:      "addr",
		Plaintext: false,
	}
}
