package main

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func getServerConfig() (map[string]string, error) {
	serverConfigs := make(map[string]string)
	file, err := os.Open("serverConfig.env")
	if err != nil {
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
		serverConfigs[keyVal[0]] = keyVal[1]
	}
	return serverConfigs, nil
}
