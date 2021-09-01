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
	ConfigFile  string   `json:"-" yaml:"-"`
	CurrentUser string   `json:"currentuser" yaml:"currentuser"`
	Users       Users    `json:"users" yaml:"users"`
	Services    Services `json:"services" yaml:"services"`
}

type User struct {
	Name    string            `json:"name" yaml:"name"`
	Headers map[string]string `json:"headers" yaml:"headers"`
}

type Header struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

type Methods struct {
	Name string `json:"name" yaml:"name"`
}

type Service struct {
	Name       string    `json:"name" yaml:"name"`
	Descriptor string    `json:"descriptor" yaml:"descriptor"`
	Methods    []Methods `json:"methods" yaml:"methods"`
	Addr       string    `json:"addr" yaml:"addr"`
	Plaintext  bool      `json:"plaintext" yaml:"plaintext"`
}

func (s Service) String() string {
	s.Descriptor = ""
	marshal, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func (c Config) SetUser(name string) (Config, error) {
	user, err := c.GetUser(name)
	if err != nil {
		return c, err
	}
	c.CurrentUser = user.Name
	return c, nil
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
		return Config{}, err
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return Config{}, err
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
		Name:       descriptors.NewServiceDescriptor(service).Command(),
		Descriptor: base64.StdEncoding.EncodeToString(fdbytes),
		Methods:    methods,
		Addr:       addr,
		Plaintext:  plaintext,
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

func DefaultUser() User {
	return User{
		Name:    "name",
		Headers: map[string]string{"key": "value"},
	}
}

func DefaultService() Service {
	return Service{
		Name:       "name",
		Descriptor: "",
		Methods: []Methods{{
			Name: "name",
		}},
		Addr:      "addr",
		Plaintext: false,
	}
}

func SetupToDataMap(v interface{}) (error, descriptors.DataMap, descriptors.DataMap) {
	var err error
	cobra.CheckErr(err)
	defaultVals, err := descriptors.NewInterfaceDataValue(v)
	flagstorer := make(descriptors.DataMap)
	cobra.CheckErr(err)
	return err, defaultVals, flagstorer
}