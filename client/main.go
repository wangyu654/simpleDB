package main

import (
	client "client/src"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Index     int      `yaml:"index"`
	Addresses []string `yaml:"addresses"`
}

func readConfig() (*Config, error) {
	var err error

	config, err := ioutil.ReadFile("../config.yaml")
	if err != nil {
		return nil, err
	}
	cfg := new(Config)
	if err = yaml.Unmarshal(config, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}
	client.Start(cfg.Addresses)
}
