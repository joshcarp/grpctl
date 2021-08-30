package grpctl

type _Config struct {
	_Somethings []_Something
}

type _Something struct {
	Name string
}

func (c _Config) Save() error {
	return nil
}

func Default_Something()_Something{
	return _Something{Name: ""}
}