package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

// 并发测试 PASS
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

	m := sync.Map{}
	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 10; i++ {
				n := rand.Intn(999)
				k := strconv.Itoa(n)
				v := strconv.Itoa(n)
				ck.Put(k, v)
				m.Store(k, v)
				log.Println("key:", k, "value:", v)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("put finished")
	m.Range(func(key, value interface{}) bool {
		expected := value.(string)
		k := key.(string)
		if v := ck.Get(k); v != expected {
			t.Error("error: expect: ", k, "but get:", v)
		} else {
			log.Println("expect:", k, "get:", v)
		}
		return true
	})

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
