package main

import (
	"fmt"
)

func GenerateIplist(n int) ([]string, []string, []string) {
	iplist := make([]string, n)
	for i := 0; i < n; i++ {
		iplist[i] = fmt.Sprintf("127.0.0.1:%d", 8000+i)
	}
	addlist := make([]string, n)
	portlist := make([]string, n)
	for i := 0; i < n; i++ {
		addlist[i] = "127.0.0.1"
		portlist[i] = fmt.Sprintf("%d", 9000+i)
	}
	return iplist, addlist, portlist
}
