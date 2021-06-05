package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Invalid argument, please specify the server type: tcp / web")
		return
	}
	serverType := args[1]
	switch serverType {
	case "tcp":
		tcpErr := startTCP()
		if tcpErr != nil {
			// todo: log error
			fmt.Println("startTcp: " + tcpErr.Error())
		}
	case "web":
		webErr := startWeb()
		if webErr != nil {
			fmt.Println("startWeb: " + webErr.Error())
		}
	case "test":
		test()
	default:
		fmt.Println("Invalid argument, the server type could only be tcp / web")
	}
}
