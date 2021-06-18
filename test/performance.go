package pers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"unsafe"

	"github.com/Amadeus-cyf/httpServer/config"
	"github.com/Amadeus-cyf/httpServer/model"
)

func Run() error {
	var (
		serverConfig map[string]string
		err          error
		url          string
	)
	if serverConfig, err = config.GetConfig(); err != nil {
		url = getUrl(serverConfig[config.WebHost], serverConfig[config.WebPort])
	}
	outputChan := make(chan string)
	errChan := make(chan error)
	for i := 0; i < 500; i++ {
		go makeApiRequests(outputChan, errChan, url, i)
	}
	for i := 0; i < 500; i++ {
		select {
		case output := <-outputChan:
			fmt.Println(output)
		case err := <-errChan:
			fmt.Println(err)
		}
	}
	return nil
}

func makeApiRequests(outputChan chan string, errChan chan error, url string, i int) {
	var err error
	username := "USER_" + strconv.Itoa(i)
	loginBody := map[string]string{"username": username, "password": "12345678"}
	var jsonVal []byte
	if jsonVal, err = json.Marshal(loginBody); err != nil {
		errChan <- err
	}
	if resp, err := http.Post(url+"/login", "application/json", bytes.NewBuffer(jsonVal)); err != nil {
		errChan <- err
	} else {
		outputChan <- fmt.Sprintf("Login Status %d, Response size: %d", resp.StatusCode, resp.ContentLength)
	}
	// update avatar
	cmd := exec.Command("curl", "-F", "data=@test.jpeg", "http://localhost:3000/update_avatar?username="+username)
	if resp, err := cmd.Output(); err != nil {
		errChan <- err
	} else {
		var httpResp model.HttpResponse
		if err = json.Unmarshal(resp, &httpResp); err != nil {
			errChan <- err
		} else {
			outputChan <- fmt.Sprintf("Update Avatar Status %d, Response size %d", httpResp.StatusCode, unsafe.Sizeof(httpResp.Body))
		}
	}
	// update nickname
	updateNicknameBody := map[string]string{"username": username, "nickname": "NICKNAME_" + strconv.Itoa((i))}
	var val []byte
	if val, err = json.Marshal(updateNicknameBody); err != nil {
		errChan <- err
	} else if resp, err := http.Post(url+"/update_nickname", "application/json", bytes.NewBuffer(val)); err != nil {
		errChan <- err
	} else {
		outputChan <- fmt.Sprintf("Update Nickname Status %d, Response size %d", resp.StatusCode, resp.ContentLength)
	}
}

func getUrl(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}
