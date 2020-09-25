package ioutil

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func FileExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func ParsePort(str string) (int, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	port := int(i)
	if port == 0 {
		port, err = FindFreePort()
		if err != nil {
			return 0, err
		}
	}
	return port, nil
}

func FindFreePort() (int, error) {
	addr := "0.0.0.0:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Failed to listen %s: %s", addr, err.Error())
		return 0, err
	}
	defer l.Close()
	return (l.Addr().(*net.TCPAddr)).Port, nil
}

type Load struct {
	Duration int     `json:"duration"`
	Average  float64 `json:"average"`
}

func LoadAvg() ([]*Load, error) {
	path := "/proc/loadavg"
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open %s: %s", path, err.Error())
		return make([]*Load, 0), nil
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Failed to read %s: %s", path, err.Error())
		return nil, err
	}
	splits := Tokenize(string(bytes))
	if len(splits) != 5 {
		err := fmt.Errorf("Unexpected %s: %s", path, string(bytes))
		log.Printf(err.Error())
		return nil, err
	}
	result := make([]*Load, 3)
	for i, _ := range result {
		result[i] = &Load{}
		result[i].Average, err = strconv.ParseFloat(splits[i], 64)
		if err != nil {
			log.Printf("Failed to parse %s: %s", path, err.Error())
			return nil, err
		}
		switch i {
		case 0:
			result[i].Duration = 60
		case 1:
			result[i].Duration = 300
		case 2:
			result[i].Duration = 900
		}
	}
	return result, nil
}

type Memory struct {
	Total int64 `json:"total"`
	Free  int64 `json:"free"`
}

func MemInfo() (*Memory, error) {
	path := "/proc/meminfo"
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open %s: %s", path, err.Error())
		return &Memory{}, nil
	}
	defer f.Close()
	info := make(map[string]int64)
	r := bufio.NewReader(f)
	for i := 0; i < 2; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Printf("Failed to read %s: %s", path, err.Error())
			return nil, err
		}
		str := strings.Trim(string(line), " \n\r")
		splits := Tokenize(str)
		if len(splits) != 3 {
			err := fmt.Errorf("Unexpected %s: %s %d", path, str, len(splits))
			log.Printf(err.Error())
			return nil, err
		}
		num, err := strconv.ParseInt(splits[1], 10, 64)
		if err != nil {
			log.Printf("Failed to parse %s: %s", path, err.Error())
			return nil, err
		}
		info[strings.Replace(splits[0], ":", "", 1)] = num
	}
	return &Memory{Total: info["MemTotal"], Free: info["MemFree"]}, nil
}

func Tokenize(str string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case ' ':
		case '\t':
		case '\r':
		case '\n':
		default:
			continue
		}
		if i > start {
			result = append(result, str[start:i])
		}
		start = i + 1
	}
	if start < len(str) {
		result = append(result, str[start:])
	}
	return result
}
