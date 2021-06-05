package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

func test() error {
	if serverConfig == nil {
		config, err := getServerConfig()
		if err != nil {
			return err
		}
		serverConfig = config
	}
	url := getUrl(serverConfig[WebHost], serverConfig[WebPort])

	before, err := cpu.Get()
	if err != nil {
		return err
	}

	for i := 0; i < 500; i++ {
		// login
		username := "USER_" + strconv.Itoa(i)
		loginBody := map[string]string{"username": username, "password": "12345678"}
		jsonVal, err := json.Marshal(loginBody)
		if err != nil {
			return err
		}
		loginResp, loginErr := http.Post(url+"/login", "application/json", bytes.NewBuffer(jsonVal))
		if loginErr != nil {
			return loginErr
		}
		fmt.Println("Login Status", loginResp.StatusCode)
		fmt.Println("Response size", loginResp.ContentLength)

		// update avatar
		cmd := exec.Command("curl", "-F", "data=@test.jpeg", "http://localhost:3000/update_avatar?username="+username)
		updateAvatarResp, cmdErr := cmd.Output()
		if cmdErr != nil {
			return cmdErr
		}
		var user User
		unmarshalErr := json.Unmarshal(updateAvatarResp, &user)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		if user.Username != "" && user.Avatar != "" {
			fmt.Println("Upate Avatar Status 200")
			fmt.Println("Response size", len(updateAvatarResp))
		}

		// update nickname
		updateNicknameBody := map[string]string{"username": username, "nickname": "NICKNAME_" + strconv.Itoa((i))}
		jsonVal2, marshalErr := json.Marshal(updateNicknameBody)
		if marshalErr != nil {
			return marshalErr
		}
		updateNicknameResp, updateErr := http.Post(url+"/update_nickname", "application/json", bytes.NewBuffer(jsonVal2))
		if updateErr != nil {
			return updateErr
		}
		fmt.Println("Update Nickname Status", updateNicknameResp.StatusCode)
		fmt.Println("Response size", updateNicknameResp.ContentLength)
	}
	after, err := cpu.Get()
	if err != nil {
		return err
	}
	mem, memErr := memory.Get()
	if memErr != nil {
		return memErr
	}
	fmt.Println("Alloc ", mem.Used)

	total := float64(after.Total - before.Total)
	fmt.Printf("cpu user: %f %%\n", float64(after.User-before.User)/total*100)
	fmt.Printf("cpu system: %f %%\n", float64(after.System-before.System)/total*100)
	fmt.Printf("cpu idle: %f %%\n", float64(after.Idle-before.Idle)/total*100)
	return nil
}

func getUrl(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}
