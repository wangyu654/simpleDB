package kvdb

import (
	"fmt"
	"strconv"
)

func (kv *KVServer) printInfo(strings ...interface{}) {
	fmt.Println(kv.PrefixString(), strings)
}

func (kv *KVServer) PrefixString() string {
	return "[" + "server id " + strconv.Itoa(kv.me)
}
