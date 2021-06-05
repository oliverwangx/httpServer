package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client

var serverConfig map[string]string

var connectionPool *ConnectionPool

func startWeb() error {
	if serverConfig == nil {
		config, err := getServerConfig()
		if err != nil {
			return err
		}
		serverConfig = config
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     serverConfig[RedisHost] + ":" + serverConfig[RedisPort],
		Password: "",
		DB:       0,
	})
	connectionPool = NewConnectionPool(1, 5, initConnectionToTcp)
	defer connectionPool.Close()
	http.HandleFunc("/login", handleLoginRequest)
	http.HandleFunc("/update_avatar", handleUpdateAvatarRequest)
	http.HandleFunc("/update_nickname", handleUpdateNickNameRequest)
	http.HandleFunc("/logout", handleLogoutRequest)
	err := http.ListenAndServe(":"+serverConfig[WebPort], nil)
	return err
}

func initConnectionToTcp() (net.Conn, error) {
	socket, err := net.DialTimeout("tcp", serverConfig[TcpHost]+":"+serverConfig[TcpPort], 1*time.Second)
	if err != nil {
		fmt.Println("Error in connection to tcp: " + err.Error())
		return nil, err
	}
	return socket, nil
}

func handleLoginRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle login request")
	conn, connErr := connectionPool.FetchConnection()
	if connErr != nil {
		fmt.Println("Connection error: " + connErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	var loginArgs LoginArgs
	err := json.NewDecoder(request.Body).Decode(&loginArgs)
	if err != nil {
		fmt.Println("handleLoginRequest, json decode error: " + err.Error())
		w.WriteHeader(ServerError)
		return
	}
	loginArgs.RequestType = Login
	argBytes, marshalErr := json.Marshal(loginArgs)
	if marshalErr != nil {
		fmt.Println("handleLoginRequest, json marshal error: " + marshalErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	sendErr := send(conn, argBytes)
	if sendErr != nil {
		fmt.Println("handleLoginRequest, send error: " + sendErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	ddlErr := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if ddlErr != nil {
		fmt.Println("handleLoginRequest, set read deadline error: " + ddlErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	receiveResponse(conn, w)
}

func handleUpdateAvatarRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle update avatar request")
	conn, connErr := connectionPool.FetchConnection()
	if connErr != nil {
		fmt.Println("Connection error: " + connErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	fmt.Println(conn.LocalAddr().String())
	username, ok := request.URL.Query()["username"]
	if !ok {
		fmt.Println("handleUpdateAvatarRequest, username not provided")
		w.WriteHeader(ServerError)
		return
	}
	if len(username) == 0 {
		fmt.Println("handleUpdateAvatarRequest, empty username")
		w.WriteHeader(ServerError)
		return
	}
	fmt.Println("Update " + username[0] + " avatar")
	file, header, formErr := request.FormFile("data")
	if formErr != nil {
		fmt.Println("handleUpdateAvatarRequest, form file fetch error: " + formErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, cpyErr := io.Copy(buf, file); cpyErr != nil {
		fmt.Println("handleUpdateAvatarRequest, copy file to buffer error: " + cpyErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	updateAvatarArgs := UpdateAvatarArgs{Username: username[0], RequestType: UpdateAvatar, Format: filepath.Ext(header.Filename), Avatar: buf.Bytes()}
	argBytes, marshalErr := json.Marshal(updateAvatarArgs)
	if marshalErr != nil {
		fmt.Println("handleUpdateAvatarRequest, json marshal error: " + marshalErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	sendErr := send(conn, argBytes)
	if sendErr != nil {
		fmt.Println("handleUpdateAvatarRequest, send error: " + sendErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	ddlErr := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if ddlErr != nil {
		fmt.Println("handleUpdateAvatarRequest, set read deadline error: " + ddlErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	receiveResponse(conn, w)
}

func handleUpdateNickNameRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle update nickname request")
	conn, connErr := connectionPool.FetchConnection()
	if connErr != nil {
		fmt.Println("Connection error: " + connErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	fmt.Println(conn.LocalAddr().String())
	var updateNickNameArgs UpdateNicknameArgs
	err := json.NewDecoder(request.Body).Decode(&updateNickNameArgs)
	if err != nil {
		fmt.Println("handleUpdateNicknameRequest, json decode error: " + err.Error())
		w.WriteHeader(ServerError)
		return
	}
	_, getCookieErr := request.Cookie(updateNickNameArgs.Username)
	if getCookieErr != nil && getCookieErr == http.ErrNoCookie {
		fmt.Println("Invalid token")
		w.WriteHeader(ServerError)
		return
	} else if err != nil {
		w.WriteHeader(ServerError)
		return
	}
	updateNickNameArgs.RequestType = UpdateNickname
	argBytes, marshalErr := json.Marshal(updateNickNameArgs)
	if marshalErr != nil {
		fmt.Println("handleUpdateNicknameRequest, json marshal error: " + marshalErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	sendErr := send(conn, argBytes)
	if sendErr != nil {
		fmt.Println("handleUpdateNicknameRequest, send error: " + sendErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	receiveResponse(conn, w)
}

func handleLogoutRequest(w http.ResponseWriter, request *http.Request) {

}

func receiveResponse(conn net.Conn, w http.ResponseWriter) {
	respBytes, receiveErr := receive(conn)
	if receiveErr != nil {
		fmt.Println("receive error: " + receiveErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	response, responseErr := handleTcpResponse(respBytes)
	if responseErr != nil || response == nil {
		fmt.Println("handle tcp response error: " + responseErr.Error())
		w.WriteHeader(ServerError)
		return
	}
	if response.Request == Login {
		// set cookie for http
		setCookie(w, response)
	}
	writeResponse(w, response)
}

func setCookie(w http.ResponseWriter, response *Response) {
	sessionInfo := response.Body.(map[string]interface{})
	if sessionInfo != nil {
		cookie := http.Cookie{Name: sessionInfo["Username"].(string), Value: sessionInfo["Token"].(string)}
		http.SetCookie(w, &cookie)
	}
}

func writeResponse(w http.ResponseWriter, response *Response) {
	w.Header().Set("Content-type", "application/json")
	bodyBytes, err := json.Marshal(response.Body)
	if err != nil {
		fmt.Println("writeResponse, json marshal error: " + err.Error())
		w.WriteHeader(ServerError)
	} else {
		w.WriteHeader(response.StatusCode)
		n, err := w.Write(bodyBytes)
		if err != nil || n < len(bodyBytes) {
			w.WriteHeader(ServerError)
		}
	}
}

func handleTcpResponse(resp []byte) (*Response, error) {
	fmt.Println("handle tcp response")
	var response Response
	err := json.Unmarshal(resp, &response)
	if err != nil {
		fmt.Println("handleTcpResponse, json unmarshal error: " + err.Error())
		return nil, err
	}
	switch response.Request {
	case Login:
		return handleLoginResponse(&response)
	case UpdateAvatar:
		return handleUpdateAvatarResponse(&response)
	case UpdateNickname:
		return handleUpdateNicknameResponse(&response)
	default:
		fmt.Println("handleTcpResponse, invalid request")
		return nil, errors.New("invalid request")
	}
}

/*
* insert user information into redis if Login Successs.
 */
func handleLoginResponse(response *Response) (*Response, error) {
	fmt.Println("Login response, status: " + strconv.Itoa(response.StatusCode))

	// TODO: use switch case
	switch response.StatusCode {
	case Success:
		sessionInfo := response.Body.(map[string]interface{})
		fmt.Println("Login Success: ", sessionInfo["Token"])
		return response, nil
	case Unauthorized:
		fmt.Println("Login: incorrect password")
		return response, nil
	case NotFound:
		fmt.Println("Login: user does not exist")
		return response, nil
	case BadRequest:
		fmt.Println("Login: bad request")
	case ServerError:
		fmt.Println("Login: server error")
		return response, nil
	}
	return nil, errors.New("invalid status code")
}

func handleUpdateAvatarResponse(response *Response) (*Response, error) {
	switch response.StatusCode {
	case Success:
		updatedUser := response.Body.(User)
		user, err := getUserDataFromCache(updatedUser.Username)
		if err != nil {
			return nil, err
		}
		user[Avatar] = updatedUser.Avatar
		updateErr := updateUserDataInCache(updatedUser.Username, user)
		if updateErr != nil {
			return nil, updateErr
		}
		fmt.Println("User avatar updated in cache")
		return response, nil
	case NotFound:
		fmt.Println("update avatar: username does not exist")
		return response, nil
	case ServerError:
		fmt.Println("update avatar: server error")
		return response, nil
	case BadRequest:
		fmt.Println("update avatar: server ServerError")
		return response, nil
	}
	return nil, errors.New("invalid status code")
}

func handleUpdateNicknameResponse(response *Response) (*Response, error) {
	switch response.StatusCode {
	case Success:
		updatedUser := response.Body.(map[string]interface{})
		user, err := getUserDataFromCache(updatedUser["Username"].(string))
		if err != nil {
			return nil, err
		}
		user[Nickname] = updatedUser["Nickname"].(string)
		updateErr := updateUserDataInCache(updatedUser["Username"].(string), user)
		if updateErr != nil {
			return nil, updateErr
		}
		fmt.Println("User nickname updated in cache")
		return response, nil
	case NotFound:
		fmt.Println("update nickname: username does not exist")
		return response, nil
	case BadRequest:
		fmt.Println("update nickanme: bad request")
		return response, nil
	case ServerError:
		fmt.Println("update nickname: username does not exist")
		return response, nil

	}
	return nil, errors.New("invalid status code")
}

// getUser as function name
func getUserDataFromCache(username string) (map[string]string, error) {
	result := rdb.HGetAll(ctx, username)
	if result == nil {
		fmt.Println("error in get user information")
		return nil, errors.New("nil result")
	} else if result.Err() != nil {
		fmt.Println("redis HGet All Error: " + result.Err().Error())
		return nil, result.Err()
	}
	updatedUser := result.Val()
	if updatedUser == nil {
		fmt.Println("user information is not stored in cache")
		return nil, errors.New("User information not found")
	}
	return updatedUser, nil
}

// updateUsser as function name
func updateUserDataInCache(username string, user map[string]string) error {
	err := rdb.HSet(ctx, username, user).Err()
	if err != nil {
		fmt.Println("error in updating user information in cache: " + err.Error())
		return err
	}
	return nil
}
