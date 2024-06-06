package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadRawIP(n int, addr string) {
	file, _ := os.OpenFile(fmt.Sprintf("rawip_%v", n), os.O_RDONLY, 0666)
	defer file.Close()
	reader := bufio.NewReader(file)

	var results []string
	// 按行处理txt
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		results = append(results, string(line))
	}
	list := make([]IPJson, n)
	for i := 0; i < n; i++ {
		id := i
		ip := results[i]
		port1 := fmt.Sprintf("%v", 8000)
		port2 := fmt.Sprintf("%v", 9000)
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

func LoadRawIP_2(n int, times int, addr string) {
	file, _ := os.OpenFile(fmt.Sprintf("rawip_%v", n), os.O_RDONLY, 0666)
	defer file.Close()
	reader := bufio.NewReader(file)

	var results []string
	// 按行处理txt
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		results = append(results, string(line))
	}
	list := make([]IPJson, times*n)
	for i := 0; i < n; i++ {
		for j := 0; j < times; j++ {
			id := times*i + j
			ip := results[i]
			port1 := fmt.Sprintf("%v", 8000+j)
			port2 := fmt.Sprintf("%v", 9000+j)
			list[times*i+j] = struct {
				ID    int
				IP    string
				Port1 string
				Port2 string
			}{ID: id, IP: ip, Port1: port1, Port2: port2}
		}

	}
	tmp3 := fmt.Sprintf("%s/iplist.json", addr)
	file3, _ := os.OpenFile(tmp3, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	encoder := json.NewEncoder(file3)
	encoder.Encode(list)
	file3.Close()
}
