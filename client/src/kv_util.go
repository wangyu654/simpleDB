package client

import (
	"fmt"
)

func (ck *Clerk) printInfo(strings ...interface{}) {
	fmt.Println(ck.PrefixString(), strings)
}

func (ck *Clerk) PrefixString() string {
	return "[" + "client" + "]"
}
