package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cellstate/errwrap"
)

var BoxDirName = ".box"

type BucketConfig struct {
	Endpoint string `json:"endpoint"`
}

type Config struct {
	Buckets []*BucketConfig `json:"buckets,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Buckets: []*BucketConfig{},
	}
}

func ReadConfig(dir string) (*Config, error) {
	fpath := filepath.Join(dir, BoxDirName, "config")
	f, err := os.Open(fpath)
	if err != nil {
		return nil, errwrap.Wrapf("Failed to read configuration file in '%s': {{err}}, is it a boxed project?", err, fpath)
	}

	conf := DefaultConfig()
	dec := json.NewDecoder(f)
	err = dec.Decode(&conf)
	if err != nil {
		return nil, errwrap.Wrapf("Failed to decode configuration file '%s': {{err}}", err, f.Name())
	}

	return conf, nil
}

func WriteConfig(dir string, conf *Config) error {
	dir = filepath.Join(dir, BoxDirName)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return errwrap.Wrapf("Failed to mkdir '%s' for config: {{err}}", err, dir)
	}

	fpath := filepath.Join(dir, "config")
	f, err := os.Create(fpath)
	if err != nil {
		return errwrap.Wrapf("Failed to create configuration file '%s': {{err}}", err, fpath)
	}

	enc := json.NewEncoder(f)
	err = enc.Encode(conf)
	if err != nil {
		return errwrap.Wrapf("Failed encode configuration '%+v': {{err}}", err, conf)
	}

	return nil
}
