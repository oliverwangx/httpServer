package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var serverConfig map[string]string

func GetConfig() (configs map[string]string, err error) {
	if serverConfig != nil {
		configs = serverConfig
		return
	}
	configs = make(map[string]string)
	var file *os.File
	if file, err = os.Open("config.env"); err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		keyVal := strings.Split(line, "=")
		if len(keyVal) != 2 {
			return nil, errors.New("incorrect config format")
		}
		configs[keyVal[0]] = keyVal[1]
	}
	serverConfig = configs
	return
}
