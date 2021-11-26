package grpctl

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type config struct {
	Entries map[string]Entry
}

type Entry struct {
	Descriptor []byte
	TTL        time.Time
}

func Load(filename string) (config, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		a := config{}.Save(filename)
		return config{}, a
	}
	var c config
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return config{}, err
	}
	c = c.Prune()
	if err = c.Save(filename); err != nil {
		return config{}, err
	}
	return c, nil
}

func (c config) AddEntry(filename string, target string, Descriptor []byte, dur time.Duration) error {
	c.Entries[target] = Entry{
		Descriptor: Descriptor,
		TTL:        time.Now().Add(dur),
	}
	return c.Save(filename)
}

func (c config) Save(filename string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, os.ModePerm)
}

func (c config) Prune() config {
	newEntries := make(map[string]Entry, len(c.Entries))
	for target, val := range c.Entries {
		if val.TTL.Before(time.Now()) {
			continue
		}
		newEntries[target] = val
	}
	c.Entries = newEntries
	return c
}
