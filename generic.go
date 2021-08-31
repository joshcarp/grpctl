//nolint
package grpctl

import (
	"encoding/json"
	"fmt"
	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"strings"
)

func Get_SomethingCommand(config _Config) *cobra.Command {
	var err error
	rootCmd := &cobra.Command{
		Use:   strings.ToLower("_Something"),
		Short: "configure _Something",
		Run:   nil,
	}
	something := Default_Something()
	cobra.CheckErr(err)
	defaultVals, err := descriptors.NewInterfaceDataValue(something)
	flagstorer := make(descriptors.DataMap)
	cobra.CheckErr(err)
	get := &cobra.Command{
		Use:   "get",
		Short: "get a _Something",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			user, err := config.Get_Something(args[0])
			cobra.CheckErr(err)
			fmt.Println(user)
		},
	}
	add := &cobra.Command{
		Use:   "add",
		Short: "add a _Something",
		Run: func(cmd *cobra.Command, args []string) {
			toJson, err := flagstorer.ToJson()
			cobra.CheckErr(err)
			cobra.CheckErr(json.Unmarshal(toJson, &something))
			config, err = config.Add_Something(something)
			cobra.CheckErr(err)
			fmt.Println(config)
			cobra.CheckErr(config.Save())
		},
	}
	del := &cobra.Command{
		Use:   "delete",
		Short: "delete a _Something",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, err = config.Delete_Something(args[0])
			cobra.CheckErr(err)
			cobra.CheckErr(config.Save())
		},
	}
	update := &cobra.Command{
		Use:   "update",
		Short: "update a _Something",
		Run: func(cmd *cobra.Command, args []string) {
			src := flagstorer.ToInterfaceMap()
			cobra.CheckErr(err)
			v, ok := flagstorer["name"]
			if !ok {
				cobra.CheckErr(InvalidArg)
			}
			context, err := config.Get_Something(v.String())
			dst, err := descriptors.ToInterfaceMap(context)
			cobra.CheckErr(err)
			allmap := descriptors.MergeInterfaceMaps(dst, src)
			cobra.CheckErr(descriptors.MapInterfaceToObject(&context, allmap))
			newcfg, err := config.Update_Something(context)
			cobra.CheckErr(err)
			cobra.CheckErr(newcfg.Save())
		},
	}
	list := &cobra.Command{
		Use:   "list",
		Short: "list all _Somethings",
		Run: func(cmd *cobra.Command, args []string) {
			for _, val := range config.List_Something() {
				marshal, err := json.Marshal(val)
				cobra.CheckErr(err)
				fmt.Println(string(marshal))
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

func (c _Config) Get_Something(name string) (_Something, error) {
	for _, e := range c._Somethings {
		if e.Name == name {
			return e, nil
		}
	}
	return _Something{}, NotFoundError
}

func (c _Config) Add_Something(s _Something) (_Config, error) {
	for _, e := range c._Somethings {
		if e.Name == s.Name {
			return _Config{}, AlreadyExists
		}
	}
	c._Somethings = append(c._Somethings, s)
	return c, nil
}

func (c _Config) Delete_Something(name string) (_Config, error) {
	for i, e := range c._Somethings {
		if e.Name == name {
			c._Somethings = append(c._Somethings[:i], c._Somethings[i+1:]...)
			return c, nil
		}
	}
	return c, NotFoundError
}

func (c _Config) Update_Something(s _Something) (_Config, error) {
	for i, e := range c._Somethings {
		if e.Name == s.Name {
			c._Somethings[i] = s
			return c, nil
		}
	}
	return c, NotFoundError
}

func (c _Config) List_Something() []_Something {
	return c._Somethings
}
