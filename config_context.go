// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

//nolint
package grpctl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
)

func GetContextCommand(config Config) *cobra.Command {
	var err error
	rootCmd := &cobra.Command{
		Use:   strings.ToLower("Context"),
		Short: "configure Context",
		Run:   nil,
	}
	something := DefaultContext()
	cobra.CheckErr(err)
	defaultVals, err := descriptors.NewInterfaceDataValue(something)
	flagstorer := make(descriptors.DataMap)
	cobra.CheckErr(err)
	get := &cobra.Command{
		Use:   "get",
		Short: "get a Context",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			user, err := config.GetContext(args[0])
			cobra.CheckErr(err)
			fmt.Println(user)
		},
	}
	add := &cobra.Command{
		Use:   "add",
		Short: "add a Context",
		Run: func(cmd *cobra.Command, args []string) {
			toJson, err := flagstorer.ToJson()
			cobra.CheckErr(err)
			cobra.CheckErr(json.Unmarshal(toJson, &something))
			config, err = config.AddContext(something)
			cobra.CheckErr(err)
			fmt.Println(config)
			cobra.CheckErr(config.Save())
		},
	}
	del := &cobra.Command{
		Use:   "delete",
		Short: "delete a Context",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, err = config.DeleteContext(args[0])
			cobra.CheckErr(err)
			cobra.CheckErr(config.Save())
		},
	}
	update := &cobra.Command{
		Use:   "update",
		Short: "update a Context",
		Run: func(cmd *cobra.Command, args []string) {
			src := flagstorer.ToInterfaceMap()
			cobra.CheckErr(err)
			v, ok := flagstorer["name"]
			if !ok {
				cobra.CheckErr(InvalidArg)
			}
			context, err := config.GetContext(v.String())
			dst, err := descriptors.ToInterfaceMap(context)
			cobra.CheckErr(err)
			allmap := descriptors.MergeInterfaceMaps(dst, src)
			cobra.CheckErr(descriptors.MapInterfaceToObject(&context, allmap))
			newcfg, err := config.UpdateContext(context)
			cobra.CheckErr(err)
			cobra.CheckErr(newcfg.Save())
		},
	}
	list := &cobra.Command{
		Use:   "list",
		Short: "list all Contexts",
		Run: func(cmd *cobra.Command, args []string) {
			for _, val := range config.ListContext() {
				fmt.Println(val)
			}
		},
	}
	for key, val := range defaultVals {
		key := key
		val := val
		flagstorer[key] = &descriptors.DataValue{Value: val.Value, Empty: true}
		update.Flags().Var(flagstorer[key], key, "")
		update.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", val)}, cobra.ShellCompDirectiveDefault
		})
		add.Flags().Var(flagstorer[key], key, "")
		add.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", val)}, cobra.ShellCompDirectiveDefault
		})
	}
	rootCmd.AddCommand(get)
	rootCmd.AddCommand(add)
	rootCmd.AddCommand(update)
	rootCmd.AddCommand(del)
	rootCmd.AddCommand(list)
	return rootCmd
}

func (c Config) GetContext(name string) (Context, error) {
	for _, e := range c.Contexts {
		if e.Name == name {
			return e, nil
		}
	}
	return Context{}, NotFoundError
}

func (c Config) AddContext(s Context) (Config, error) {
	for _, e := range c.Contexts {
		if e.Name == s.Name {
			return Config{}, AlreadyExists
		}
	}
	c.Contexts = append(c.Contexts, s)
	return c, nil
}

func (c Config) DeleteContext(name string) (Config, error) {
	for i, e := range c.Contexts {
		if e.Name == name {
			c.Contexts = append(c.Contexts[:i], c.Contexts[i+1:]...)
			return c, nil
		}
	}
	return c, NotFoundError
}

func (c Config) UpdateContext(s Context) (Config, error) {
	for i, e := range c.Contexts {
		if e.Name == s.Name {
			c.Contexts[i] = s
			return c, nil
		}
	}
	return c, NotFoundError
}

func (c Config) ListContext() []Context {
	return c.Contexts
}
