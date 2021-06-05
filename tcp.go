package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func startTCP() error {
	database, err := sql.Open("mysql", "root:12345678@/ENTRY_TASK")
	if err != nil {
		fmt.Println("database connection error: " + err.Error())
		return err
	}
	db = database
	if serverConfig == nil {
		config, err := getServerConfig()
		if err != nil {
			return err
		}
		serverConfig = config
	}
	resgisterHttpUrl()
	listenErr := listen()
	return listenErr
}

func listen() error {
	fmt.Println("Listening...")
	port, portErr := strconv.Atoi(serverConfig[TcpPort])
	if portErr != nil {
		return portErr
	}
	addrInfo := net.TCPAddr{
		IP:   net.ParseIP(serverConfig[TcpHost]),
		Port: port,
	}
	err := createDirIfNotExist(AvatarDirectory)
	if err != nil {
		return err
	}
	ln, err := net.ListenTCP("tcp", &addrInfo)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		fmt.Println(conn.RemoteAddr().String())
		if err != nil {
			fmt.Println("Accept: " + err.Error())
			continue
		}
		go processRequest(conn)
	}
}

func processRequest(conn net.Conn) error {
	for {
		req, err := receive(conn)
		defer conn.Close()
		if err != nil {
			fmt.Println("process request error: " + err.Error())
			return err
		}
		response, requestErr := handleRequest(req)
		if requestErr != nil {
			fmt.Println(requestErr.Error())
			return requestErr
		}
		responseBytes, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			fmt.Println("process request error: " + marshalErr.Error())
			return err
		}
		sendErr := send(conn, responseBytes)
		if sendErr != nil {
			fmt.Println("process request error: " + sendErr.Error())
			return err
		}
	}
}

func handleRequest(request []byte) (*Response, error) {
	var params CoreParams
	err := json.Unmarshal(request, &params)
	if err != nil {
		return nil, err
	}
	fmt.Println("Handle Request: " + params.RequestType)
	switch params.RequestType {
	case Login:
		params, err := processLoginArgs(request)
		if err != nil {
			return nil, err
		}
		response, loginErr := login(params)
		if loginErr != nil {
			return &Response{ServerError, Login, loginErr.Error(), nil}, nil
		}
		return response, nil
	case UpdateAvatar:
		params, err := processUpdateAvatarArgs(request)
		if err != nil {
			return nil, err
		}
		response, updateAvatarErr := updateAvatar(params)
		if updateAvatarErr != nil {
			return &Response{ServerError, UpdateAvatar, updateAvatarErr.Error(), nil}, nil
		}
		return response, nil
	case UpdateNickname:
		params, err := processUpdateNicknameArgs(request)
		if err != nil {
			return nil, err
		}
		response, updateNicknameErr := updateNickname(params)
		if updateNicknameErr != nil {
			return &Response{ServerError, UpdateNickname, updateNicknameErr.Error(), nil}, nil
		}
		return response, nil
	default:
		return nil, errors.New("invalid command")
	}
}

func processLoginArgs(args []byte) (*LoginParams, error) {
	var params LoginParams
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func login(params *LoginParams) (*Response, error) {
	fmt.Println("Login")
	var user User
	queryErr := db.QueryRow("SELECT username, password, avatar, nickname FROM User WHERE username = ?", params.Username).Scan(&user.Username, &user.Password, &user.Avatar, &user.Nickname)
	if queryErr == sql.ErrNoRows {
		fmt.Println("login: " + queryErr.Error())
		return &Response{NotFound, Login, queryErr.Error(), User{}}, nil
	} else if queryErr != nil {
		fmt.Println("query error in login: " + queryErr.Error())
		return nil, queryErr
	}
	if user.Password != params.Password {
		return &Response{Unauthorized, Login, "INCORRECT PASSWORD", nil}, nil
	}
	// generate a token back to user
	token := hash(user.Username)
	fmt.Println("token generated: ", token)
	result, insertErr := db.Exec("INSERT INTO Session (Username, Token) VALUES (?, ?)", user.Username, token)
	if insertErr != nil {
		fmt.Println("Insert error" + insertErr.Error())
		return nil, insertErr
	}
	num, rowsAffectedErr := result.RowsAffected()
	if rowsAffectedErr != nil {
		return nil, insertErr
	}
	if num == 0 {
		return nil, errors.New("insert failed")
	}
	return &Response{Success, Login, "Login Success", Session{user.Username, strconv.FormatUint(token, 16)}}, nil
}

func processUpdateAvatarArgs(args []byte) (*UpdateAvatarParams, error) {
	var params UpdateAvatarParams
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func updateAvatar(params *UpdateAvatarParams) (*Response, error) {
	fmt.Println("Update avatar")
	if params.Username == "" || params.Avatar == nil {
		return &Response{BadRequest, UpdateAvatar, "No Username", User{}}, nil
	}
	var user User
	queryErr := db.QueryRow("SELECT username, nickname, avatar FROM User WHERE username = ?", params.Username).Scan(&user.Username, &user.Nickname, &user.Avatar)
	if queryErr == sql.ErrNoRows {
		return &Response{NotFound, UpdateAvatar, "User not found", User{}}, nil
	} else if queryErr != nil {
		return nil, queryErr
	}
	avatarDir := filepath.Join(AvatarDirectory, params.Username)
	createDirErr := createDirIfNotExist(avatarDir)
	if createDirErr != nil {
		return nil, createDirErr
	}
	path := filepath.Join(avatarDir, "profile"+params.Format)
	writeErr := ioutil.WriteFile(path, params.Avatar, 0644)
	if writeErr != nil {
		return nil, writeErr
	}
	url := fmt.Sprintf("%s:%s/%s", serverConfig[TcpHost], serverConfig[AvatarPort], path)
	_, updateErr := db.Exec("UPDATE User Set avatar = ? WHERE username = ?", url, params.Username)
	if updateErr != nil {
		return nil, updateErr
	}
	user.Avatar = url
	return &Response{Success, UpdateAvatar, "Avatar updated", user}, nil
}

func processUpdateNicknameArgs(args []byte) (*UpdateNicknameParams, error) {
	var params UpdateNicknameParams
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func updateNickname(params *UpdateNicknameParams) (*Response, error) {
	fmt.Println("Update nickname")
	if params.Username == "" {
		return &Response{BadRequest, UpdateNickname, "No username", User{}}, nil
	}
	var user User
	queryErr := db.QueryRow("SELECT username, nickname, avatar FROM User Where username = ?", params.Username).Scan(&user.Username, &user.Nickname, &user.Avatar)
	if queryErr == sql.ErrNoRows {
		fmt.Println("User does not exist")
		return &Response{NotFound, UpdateNickname, "Username not found", User{}}, nil
	} else if queryErr != nil {
		return nil, queryErr
	}
	_, updateErr := db.Exec("UPDATE User Set nickname = ? WHERE username = ?", params.Nickname, params.Username)
	if updateErr != nil {
		return nil, updateErr
	}
	user.Nickname = params.Nickname
	return &Response{Success, UpdateNickname, "Nickname updated", user}, nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err2 := os.MkdirAll(dir, os.ModePerm)
		if err2 != nil {
			return err2
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func resgisterHttpUrl() {
	http.HandleFunc("/", avatarViewHandler)
	port := ":" + serverConfig[AvatarPort]
	go http.ListenAndServe(port, nil)
}

func avatarViewHandler(w http.ResponseWriter, request *http.Request) {
	path := request.URL.Path[1:]
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("error in reading directory " + path)
	} else {
		n, writeErr := w.Write(bytes)
		if writeErr != nil || n < len(bytes) {
			w.WriteHeader(ServerError)
		}
	}
}
