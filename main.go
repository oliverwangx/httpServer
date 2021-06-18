package main

import (
	"fmt"
	"os"

	httpServer "github.com/Amadeus-cyf/httpServer/http"
	"github.com/Amadeus-cyf/httpServer/tcp"
	test "github.com/Amadeus-cyf/httpServer/test"
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
		err := tcp.StartTCP()
		if err != nil {
			fmt.Println("startTcp: " + err.Error())
		}
	case "http":
		err := httpServer.StartHttp()
		if err != nil {
			fmt.Println("startWeb: " + err.Error())
		}
	case "test":
		test.Run()
	default:
		fmt.Println("Invalid argument, the server type could only be tcp / web")
	}
}
