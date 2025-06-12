package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/QinYuuuu/abvss/harts"
)

func main() {
	var (
		logLevel  string
		logFormat string
	)
	flag.StringVar(&logLevel, "log", "info", "set log level to debug, info, warn, error or fatal (case-insensitive). default is INFO")
	flag.StringVar(&logFormat, "format", "json", "set log format to json or text. default is json")
	flag.Parse()
	setupLog(logLevel, logFormat)
	configByte, err := os.ReadFile("test_config.json")
	if err != nil {
		panic(err)
	}
	var testConfig Config
	err = json.Unmarshal(configByte, &testConfig)
	if err != nil {
		panic(err)
	}
	path := testConfig.OutputPath
	exist, err := pathExists(path)
	if err != nil {
		panic(err)
	}
	if !exist {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	record := Record{}
	for i, testCase := range testConfig.TestCase {
		record.TestCase = append(record.TestCase, struct {
			TestNum   int64         `json:"test_num"`
			NodeNum   int64         "json:\"node_num\""
			Threshold int64         "json:\"threshold\""
			BatchSize int64         "json:\"batch_size\""
			Usage     []harts.Usage "json:\"usage\""
		}{
			TestNum:   testCase.TestNum,
			NodeNum:   testCase.NodeNum,
			Threshold: testCase.Threshold,
			BatchSize: testCase.BatchSize,
		})
		for j := range testCase.TestNum {
			slog.Info(fmt.Sprintf("test:[node %v][thre %v][batch %v][times %v]", testCase.NodeNum, testCase.Threshold, testCase.BatchSize, j+1))
			usage := harts.InitLocalDKG(testCase.NodeNum, testCase.Threshold, testCase.BatchSize)
			record.TestCase[i].Usage = append(record.TestCase[i].Usage, *usage)
		}

		// filename := fmt.Sprintf("node_%v_threshold_%v_batch_%v", testCase.NodeNum, testCase.Threshold, testCase.BatchSize)
	}
	recordBytes, err := json.Marshal(record)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(path+"/test.json", recordBytes, os.ModePerm)
	if err != nil {
		panic(err)
	}
	fmt.Println("Press any key to exit...")
	b := make([]byte, 1)
	os.Stdin.Read(b)
}

type Config struct {
	TestCase []struct {
		TestNum   int64 `json:"test_num"`
		NodeNum   int64 `json:"node_num"`
		Threshold int64 `json:"threshold"`
		BatchSize int64 `json:"batch_size"`
	} `json:"test_case"`
	OutputPath string `json:"output_path"`
}

type Record struct {
	TestCase []struct {
		TestNum   int64         `json:"test_num"`
		NodeNum   int64         `json:"node_num"`
		Threshold int64         `json:"threshold"`
		BatchSize int64         `json:"batch_size"`
		Usage     []harts.Usage `json:"usage"`
	} `json:"test_case"`
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func setupLog(lvl, format string) {
	logLevel := slog.LevelInfo.Level()
	var logger *slog.Logger
	if len(lvl) > 0 {
		err := logLevel.UnmarshalText([]byte(lvl))
		// logLevel not change if unmarshall failed
		if err != nil {
			fmt.Println("input invalid log level, use default log level INFO")
		}
	}
	// TODO:log source file position
	opt := &slog.HandlerOptions{AddSource: false, Level: logLevel}
	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opt)
	default:
		handler = slog.NewTextHandler(os.Stdout, opt)
	}
	fmt.Printf("init logger, level: %s, format: %s\n", logLevel.String(), format)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}
