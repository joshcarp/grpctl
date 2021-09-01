//nolint
package grpctl

import (
	"encoding/json"
	"fmt"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
)

func Get_SomethingCommands(config _Config) []*cobra.Command {
	return []*cobra.Command{
		Get_SomethingGetCommand(config),
		Get_SomethingAddCommand(config),
		Get_SomethingDeleteCommand(config),
		Get_SomethingClearCommand(config),
		Get_SomethingUpdateCommand(config),
		Get_SomethingListCommand(config),
	}
}

func Get_SomethingGetCommand(config _Config) *cobra.Command {
	get := &cobra.Command{
		Use:               "get",
		Short:             "get a _Something",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.MinimumNArgs(1),
		ValidArgs:         config._Somethings.Names(),
		Run: func(cmd *cobra.Command, args []string) {
			user, err := config.Get_Something(args[0])
			cobra.CheckErr(err)
			fmt.Println(user)
		},
	}
	return get
}

func Get_SomethingAddCommand(config _Config) *cobra.Command {
	something := Default_Something()
	err, defaultVals, flagstorer := SetupToDataMap(&something)
	cobra.CheckErr(err)
	add := &cobra.Command{
		Use:               "add",
		Short:             "add a _Something",
		ValidArgsFunction: cobra.NoFileCompletions,
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
	flagCompletion(defaultVals, flagstorer, add)
	return add
}

func Get_SomethingDeleteCommand(config _Config) *cobra.Command {
	return &cobra.Command{
		Use:               "delete",
		Short:             "delete a _Something",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		ValidArgs:         config._Somethings.Names(),
		Run: func(cmd *cobra.Command, args []string) {
			config, err := config.Delete_Something(args[0])
			cobra.CheckErr(err)
			cobra.CheckErr(config.Save())
		},
	}
}

func Get_SomethingClearCommand(config _Config) *cobra.Command {
	return &cobra.Command{
		Use:               "clear",
		Short:             "clear all _Somethings",
		Args:              cobra.ExactArgs(0),
		ValidArgsFunction: cobra.NoFileCompletions,
		ValidArgs:         config._Somethings.Names(),
		Run: func(cmd *cobra.Command, args []string) {
			config._Somethings = nil
			cobra.CheckErr(config.Save())
		},
	}
}

func Get_SomethingUpdateCommand(config _Config) *cobra.Command {
	something := Default_Something()
	err, defaultVals, flagstorer := SetupToDataMap(&something)
	update := &cobra.Command{
		Use:               "update",
		Short:             "update a _Something",
		Args:              cobra.ExactArgs(0),
		ValidArgsFunction: cobra.NoFileCompletions,
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
	flagCompletion(defaultVals, flagstorer, update)
	return update
}

func Get_SomethingListCommand(config _Config) *cobra.Command {
	return &cobra.Command{
		Use:               "list",
		Short:             "list all _Somethings",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			for _, val := range config.List_Something() {
				fmt.Println(val)
			}
		},
	}
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

type _Somethings []_Something

func (s _Somethings) Names() []string {
	var names []string
	for _, user := range s {
		names = append(names, user.Name)
	}
	return names
}
