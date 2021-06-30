package client

import (
	"fmt"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestPut(t *testing.T) {
	var err error

	config, err := ioutil.ReadFile("../../config.yaml")
	if err != nil {
		return
	}
	cfg := new(Config)
	if err = yaml.Unmarshal(config, cfg); err != nil {
		return
	}
	ck := Start(cfg.Addresses)
	ck.Put("3", "1")
}

func TestGet(t *testing.T) {
	var err error

	config, err := ioutil.ReadFile("../../config.yaml")
	if err != nil {
		return
	}
	cfg := new(Config)
	if err = yaml.Unmarshal(config, cfg); err != nil {
		return
	}
	ck := Start(cfg.Addresses)
	fmt.Println("res:", ck.Get("3"))
}
