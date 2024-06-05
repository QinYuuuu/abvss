package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func SHRawIP(n int) {
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
	tmp3 := fmt.Sprintf("ship")
	file3, _ := os.OpenFile(tmp3, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	for i := 0; i < len(results); i++ {
		tmp := fmt.Sprintf("[%v]='%v'\n", i, results[i])
		file3.WriteString(tmp)
	}
	file3.Close()
}
