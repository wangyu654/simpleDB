package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

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

	rand.Seed(time.Now().Unix())

	m := map[int]int{}
	for i := 0; i < 20; i++ {
		n := rand.Intn(999)
		k := strconv.Itoa(n)
		v := strconv.Itoa(n)
		ck.Put(k, v)
		m[n] = n
		log.Println("key:", k, "value:", v)
	}
	log.Println("put finished")
	for k, value := range m {
		expected := strconv.Itoa(value)
		if v := ck.Get(strconv.Itoa(k)); v != expected {
			t.Error("error: expect: ", k, "but get:", v)
		} else {
			log.Println("expect:", k, "get:", v)
		}
	}
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
