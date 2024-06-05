package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type IPJson struct {
	ID    int
	IP    string
	Port1 string
	Port2 string
}

func GenerateIPList_Local(n int, addr string) {
	list := make([]IPJson, n)
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
	tmp3 := fmt.Sprintf("%s/iplist_local.json", addr)
	file3, _ := os.OpenFile(tmp3, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	encoder := json.NewEncoder(file3)
	encoder.Encode(list)
	file3.Close()
}

func LoadIPList_aws(n int, addr string) ([]string, []string, []string) {
	list := make([]IPJson, n)
	tmp3 := fmt.Sprintf("%s/iplist.json", addr)
	input, _ := os.ReadFile(tmp3)
	json.Unmarshal(input, &list)
	iplist1 := make([]string, n)
	for i := 0; i < n; i++ {
		iplist1[i] = "127.0.0.1" + ":" + list[i].Port1
	}
	iplist2 := make([]string, n)
	for i := 0; i < n; i++ {
		iplist2[i] = list[i].IP
	}
	iplist3 := make([]string, n)
	for i := 0; i < n; i++ {
		iplist3[i] = list[i].Port2
	}
	return iplist1, iplist2, iplist3
}

func LoadIPList_Local(n int, addr string) ([]string, []string, []string) {
	list := make([]IPJson, n)
	tmp3 := fmt.Sprintf("%s/iplist_local.json", addr)
	input, _ := os.ReadFile(tmp3)
	json.Unmarshal(input, &list)
	//fmt.Println(list)
	iplist1 := make([]string, n)
	for i := 0; i < n; i++ {
		iplist1[i] = list[i].IP + ":" + list[i].Port1
	}
	iplist2 := make([]string, n)
	for i := 0; i < n; i++ {
		iplist2[i] = list[i].IP
	}
	iplist3 := make([]string, n)
	for i := 0; i < n; i++ {
		iplist3[i] = list[i].Port2
	}
	return iplist1, iplist2, iplist3
}
