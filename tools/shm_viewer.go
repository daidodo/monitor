package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/daidodo/overlord/inner"
)

type attrs []uint32

func (a *attrs) String() string {
	return fmt.Sprint(*a)
}

func (a *attrs) Set(s string) error {
	for _, v := range strings.Split(s, ",") {
		if i, e := strconv.ParseUint(v, 10, 32); e != nil {
			return e
		} else {
			*a = append(*a, uint32(i))
		}
	}
	return nil
}

func found(as attrs, a uint32) bool {
	for _, i := range as {
		if i == a {
			return true
		}
	}
	return false
}

func main() {
	ns, err := inner.Attach(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot attach overlord shm: %v\n", err)
		return
	}
	var as attrs
	flag.Var(&as, "a", "comma-separated list of attributes, e.g., 1,2,3")
	flag.Parse()
	for _, n := range ns {
		if n.Attr == 0 {
			continue
		}
		if len(as) == 0 || found(as, n.Attr) {
			fmt.Println(n.Attr, ":", n.Value)
		}
	}
}
