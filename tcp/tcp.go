package tcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/Amadeus-cyf/httpServer/config"
	"github.com/Amadeus-cyf/httpServer/consts"
	requestType "github.com/Amadeus-cyf/httpServer/consts/request"
	statusCode "github.com/Amadeus-cyf/httpServer/consts/status_code"
	"github.com/Amadeus-cyf/httpServer/model"
	"github.com/Amadeus-cyf/httpServer/socketIO"
	"github.com/Amadeus-cyf/httpServer/store"
	"github.com/Amadeus-cyf/httpServer/utils"
	_ "github.com/go-sql-driver/mysql"
)

var serverConfig map[string]string
var ctx context.Context
var dataStore *store.DataStore

func StartTCP() (err error) {
	if serverConfig, err = config.GetConfig(); err != nil {
		return
	}
	dataStore = new(store.DataStore)
	if err = dataStore.Init(); err != nil {
		return
	}
	ctx = context.Background()
	resgisterHttpUrl()
	err = listen()
	return
}

func listen() (err error) {
	fmt.Println("Listening...")
	var port int
	if port, err = strconv.Atoi(serverConfig[config.TcpPort]); err != nil {
		return
	}
	addrInfo := net.TCPAddr{
		IP:   net.ParseIP(serverConfig[config.TcpHost]),
		Port: port,
	}
	if err = utils.CreateDirIfNotExist(consts.AvatarDirectory); err != nil {
		return
	}
	var ln *net.TCPListener
	if ln, err = net.ListenTCP("tcp", &addrInfo); err != nil {
		return
	}
	for {
		var conn net.Conn
		if conn, err = ln.Accept(); err != nil {
			fmt.Println("Accept: " + err.Error())
			continue
		}
		go receiveRequest(conn)
	}
}

func receiveRequest(conn net.Conn) {
	defer conn.Close()
	for {
		var req []byte
		var err error
		var response model.Response
		if req, err = socketIO.Receive(conn); err != nil {
			fmt.Println("process request error: " + err.Error())
			response = model.NewErrResponse()
		}
		if response, err = handleRequest(req); err != nil {
			fmt.Println("handle request error: " + err.Error())
			response = model.NewErrResponse()
			return
		}
		fmt.Println(response)
		var responseBytes []byte
		if responseBytes, err = json.Marshal(response); err != nil {
			fmt.Println("process request error: " + err.Error())
			return
		}
		if err = socketIO.Send(conn, responseBytes); err != nil {
			fmt.Println("process request error: " + err.Error())
			return
		}
	}
}

func handleRequest(request []byte) (resp model.Response, err error) {
	var params model.BaseParams
	err = json.Unmarshal(request, &params)
	if err != nil {
		return nil, err
	}
	fmt.Println("Handle Request: " + params.RequestType)
	switch params.RequestType {
	case requestType.Login:
		loginParams := new(model.LoginParams)
		if err = json.Unmarshal(request, loginParams); err != nil {
			return
		}
		resp, err = login(loginParams)
		if err != nil && resp.GetRequestType() == "" {
			resp = model.CastToTcpLoginResp(statusCode.ServerError, requestType.Login, model.User{}, "")
			err = nil
		}
	case requestType.UpdateAvatar:
		updateAvatarParams := new(model.UpdateAvatarParams)
		if err = json.Unmarshal(request, updateAvatarParams); err != nil {
			return
		}
		resp, err = updateAvatar(updateAvatarParams)
		if err != nil && resp.GetRequestType() == "" {
			resp = model.CastToTcpUpdateResp(statusCode.ServerError, requestType.UpdateAvatar, model.User{})
		}
		err = nil
	case requestType.UpdateNickname:
		updateNicknameParams := new(model.UpdateNicknameParams)
		if err = json.Unmarshal(request, updateNicknameParams); err != nil {
			return
		}
		resp, err = updateNickname(updateNicknameParams)
		if err != nil && resp.GetRequestType() == "" {
			resp = model.CastToTcpUpdateResp(statusCode.ServerError, requestType.UpdateNickname, model.User{})
		}
		err = nil
	default:
		err = errors.New("invalid command")
	}
	return
}

func login(params *model.LoginParams) (resp model.TcpLoginResponse, err error) {
	fmt.Println("Login")
	var user *model.User
	if user, err = dataStore.GetUserByUsername(ctx, params.Username); err != nil {
		if err == sql.ErrNoRows {
			resp = model.CastToTcpLoginResp(statusCode.NotFound, requestType.Login, model.User{}, "")
			err = nil
		}
		return
	}
	fmt.Println("User", user)
	if user.Password != params.Password {
		resp = model.CastToTcpLoginResp(statusCode.Unauthorized, requestType.Login, model.User{Username: params.Username}, "")
		return
	}
	// generate a session token
	token := strconv.FormatUint(utils.Hash(user.Username), 16)
	fmt.Println("token generated: ", token, "for ", user.Username)
	if err = dataStore.SetUserSession(ctx, user.Username, token); err != nil {
		fmt.Println("Set user session error", err.Error())
		return
	}
	user.Password = ""
	resp = model.CastToTcpLoginResp(statusCode.Success, requestType.Login, *user, token)
	return
}

func updateAvatar(params *model.UpdateAvatarParams) (resp model.TcpUpdateResponse, err error) {
	fmt.Println("Update avatar")
	if params.Username == "" || params.Avatar == nil {
		resp = model.CastToTcpUpdateResp(statusCode.BadRequest, requestType.UpdateAvatar, model.User{})
		return
	}
	var token string
	if token, err = dataStore.GetUserSession(ctx, params.Username); err != nil || token != params.Token {
		resp = model.CastToTcpUpdateResp(statusCode.Unauthorized, requestType.UpdateAvatar, model.User{})
		return
	}
	var user *model.User
	if user, err = dataStore.GetUserByUsername(ctx, params.Username); err != nil {
		if err == sql.ErrNoRows {
			resp = model.CastToTcpUpdateResp(statusCode.NotFound, requestType.UpdateAvatar, model.User{})
		}
		return
	}
	avatarDir := filepath.Join(consts.AvatarDirectory, params.Username)
	if err = utils.CreateDirIfNotExist(avatarDir); err != nil {
		return
	}
	path := filepath.Join(avatarDir, "profile"+params.Format)
	if err = ioutil.WriteFile(path, params.Avatar, 0644); err != nil {
		return
	}
	url := fmt.Sprintf("%s:%s/%s", serverConfig[config.TcpHost], serverConfig[config.AvatarPort], path)
	if err = dataStore.UpdateUserAvatar(ctx, params.Username, url); err != nil {
		return
	}
	user.Avatar = url
	user.Password = ""
	resp = model.CastToTcpUpdateResp(statusCode.Success, requestType.UpdateAvatar, *user)
	return
}

func updateNickname(params *model.UpdateNicknameParams) (resp model.TcpUpdateResponse, err error) {
	fmt.Println("Update nickname")
	if params.Username == "" {
		resp = model.CastToTcpUpdateResp(statusCode.BadRequest, requestType.UpdateNickname, model.User{})
		return
	}
	var user *model.User
	if user, err = dataStore.GetUserByUsername(ctx, params.Username); err != nil {
		if err == sql.ErrNoRows {
			resp = model.CastToTcpUpdateResp(statusCode.NotFound, requestType.UpdateNickname, model.User{})
		}
		return
	}
	if err = dataStore.UpdateUserNickname(ctx, params.Username, params.Nickname); err != nil {
		return
	}
	user.Nickname = params.Nickname
	user.Password = ""
	resp = model.CastToTcpUpdateResp(statusCode.Success, requestType.UpdateNickname, *user)
	return
}

func resgisterHttpUrl() {
	http.HandleFunc("/", avatarViewHandler)
	port := ":" + serverConfig[config.AvatarPort]
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
			w.WriteHeader(statusCode.ServerError)
		}
	}
}
