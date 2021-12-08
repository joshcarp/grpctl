package grpctl

import (
	"os"
	"path"
	"time"

	"github.com/joshcarp/grpctl/internal/grpc"

	"github.com/spf13/cobra"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func reflectFileDesc(flags []string) ([]protoreflect.FileDescriptor, error) {
	cmd := cobra.Command{
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	err := persistentFlags(&cmd, "")
	if err != nil {
		return nil, err
	}

	if len(flags) > 0 && flags[0] == "__complete" {
		flags = flags[1:]
	}
	cmd.SetArgs(flags)
	var fds []protoreflect.FileDescriptor
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("address")
		if err != nil {
			return err
		}
		if addr == "" {
			return nil
		}
		cfgFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		plaintext, err := cmd.Flags().GetBool("plaintext")
		if err != nil {
			return err
		}
		if cfgFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			cfgFile = path.Join(home, ".grpctl.yaml")
			if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
				err = config{}.save(cfgFile)
				if err != nil {
					return err
				}
			}
		}
		cfg, err := loadConfig(cfgFile)
		if err != nil {
			return err
		}
		desc, err := cfg.Entries[addr].decodeDescriptor()
		if err != nil {
			return err
		}
		if len(desc) != 0 {
			spb := &descriptorpb.FileDescriptorSet{}
			err = proto.Unmarshal(desc, spb)
			if err != nil {
				return err
			}
			fds, err = grpc.ConvertToProtoReflectDesc(spb)
			if err != nil {
				return err
			}
			return nil
		}
		conn, err := grpc.Setup(cmd.Context(), plaintext, addr)
		if err != nil {
			return err
		}
		fdset, err := grpc.Reflect(conn)
		if err != nil {
			return err
		}
		fds, err = grpc.ConvertToProtoReflectDesc(fdset)
		if err != nil {
			return err
		}
		b, err := proto.Marshal(fdset)
		if err != nil {
			return err
		}
		if err := cfg.add(cfgFile, addr, b, time.Minute*15); err != nil {
			return err
		}
		return nil
	}

	err = cmd.Execute()
	return fds, err
}
