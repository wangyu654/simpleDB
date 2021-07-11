package bptree

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	var (
		tree *Tree
		err  error
	)
	if tree, err = NewTree("./data.wy"); err != nil {
		t.Fatal(err)
	}

	tree.Insert(12, "cl")
}

func BenchmarkGet(b *testing.B) {

}

func BenchmarkPut(b *testing.B) {
	b.StopTimer()

	var (
		tree *Tree
		err  error
	)

	rand.Seed(time.Now().Unix())
	if tree, err = NewTree("./data.wy"); err != nil {
		b.Fatal(err)
	}

	m := sync.Map{}
	wg := sync.WaitGroup{}
	b.StartTimer()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100000; i++ {
				n := rand.Intn(999999)
				k := uint64(n)
				v := strconv.Itoa(n)
				err = tree.Insert(k, v)
				if err != nil {
					m.Store(k, v)
					log.Println("key:", k, "value:", v)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	b.StopTimer()

	log.Println("put finished")
	// m.Range(func(key, value interface{}) bool {
	// 	expected := value.(string)
	// 	k := key.(uint64)
	// 	if v, err := tree.Find(k); err != nil || v != expected {
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		} else {
	// 			b.Error("error: expect: ", k, "but get:", v)
	// 			log.Println("expect:", k, "but get:", v)
	// 		}
	// 	} else {
	// 		log.Println("expect:", k, "get:", v)
	// 	}
	// 	return true
	// })

}

func TestGet(t *testing.T) {
	var (
		tree *Tree
		err  error
	)
	if tree, err = NewTree("./data.wy"); err != nil {
		t.Fatal(err)
	}
	fmt.Println(tree.Find(12))

}
