package grpctl

import (
	"encoding/base64"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type config struct {
	Entries map[string]entry
}

type entry struct {
	Descriptor string
	Expiry     time.Time
}

func (e entry) decodeDescriptor() ([]byte, error) {
	return base64.StdEncoding.DecodeString(e.Descriptor)
}

func loadConfig(filename string) (config, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		a := config{}.save(filename)
		return config{}, a
	}
	var c config
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return config{}, err
	}
	c = c.prune()
	if err = c.save(filename); err != nil {
		return config{}, err
	}
	return c, nil
}

func (c config) add(filename string, target string, Descriptor []byte, dur time.Duration) error {
	c.Entries[target] = entry{
		Descriptor: base64.StdEncoding.EncodeToString(Descriptor),
		Expiry:     time.Now().Add(dur),
	}
	return c.save(filename)
}

func (c config) save(filename string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, os.ModePerm)
}

func (c config) prune() config {
	newEntries := make(map[string]entry, len(c.Entries))
	for target, val := range c.Entries {
		if val.Expiry.Before(time.Now()) {
			continue
		}
		newEntries[target] = val
	}
	c.Entries = newEntries
	return c
}
