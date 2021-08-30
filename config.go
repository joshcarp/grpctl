package grpctl

import (
	"encoding/base64"
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
	ConfigFile     string    `yaml:"-"`
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
	Users          []User    `yaml:"users"`
	Services       []Service `yaml:"services"`
}

type User struct {
	Name    string   `yaml:"name"`
	Headers []Header `yaml:"headers"`
}

type Context struct {
	Name            string `yaml:"name"`
	UserName        string `yaml:"user"`
	EnvironmentName string `yaml:"env"`
	User            User   `yaml:"-"`
}

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Methods struct {
	Name string `yaml:"name"`
}

type Environment struct {
	Name      string `yaml:"name"`
	Addr      string `yaml:"addr"`
	Plaintext bool   `yaml:"plaintext"`
}

type Service struct {
	Parent       *Config       `yaml:"-"`
	Environments []Environment `yaml:"environments"`
	Name         string        `yaml:"name"`
	Descriptor   string        `yaml:"descriptor"`
	Methods      []Methods     `yaml:"methods"`
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
	return c.Parent.Save()
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

func GetCurrentContext(cfg Config) (Context, error) {
	for _, e := range cfg.Contexts {
		if e.Name == cfg.CurrentContext {
			return e, nil
		}
	}
	return Context{}, NotFoundError
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
		Name:            "",
		UserName:        "",
		EnvironmentName: "",
		User: User{
			Name: "",
			Headers: []Header{
				{
					Key:   "",
					Value: "",
				},
			},
		},
	}
}

func DefaultUser() User {
	return User{
		Name:    "",
		Headers: []Header{{
			Key:   "",
			Value: "",
		}},
	}
}

func DefaultService() Service {
	return Service{
		Environments: []Environment{{
			Name:      "",
			Addr:      "",
			Plaintext: false,
		}},
		Name:       "",
		Descriptor: "",
		Methods: []Methods{{
			Name: "",
		}},
	}
}

func DefaultEnvironment() Environment {
	return Environment{
		Name:      "",
		Addr:      "",
		Plaintext: false,
	}
}
