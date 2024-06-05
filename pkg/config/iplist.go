package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func GenerateIPList_Local(n int, addr string) {
	list := make([]struct {
		ID    int
		IP    string
		Port1 string
		Port2 string
	}, n)
	for i := 0; i < n; i++ {
		id := i
		ip := "127.0.0.1"
		port1 := fmt.Sprintf("%v", 8000+i)
		port2 := fmt.Sprintf("%v", 9000+i)
		list[i] = struct {
			ID    int
			IP    string
			Port1 string
			Port2 string
		}{ID: id, IP: ip, Port1: port1, Port2: port2}
	}
	tmp3 := fmt.Sprintf("%s/iplist.json", addr)
	file3, _ := os.OpenFile(tmp3, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	encoder := json.NewEncoder(file3)
	encoder.Encode(list)
	file3.Close()
}
