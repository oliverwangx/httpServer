package httpServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Amadeus-cyf/httpServer/config"
	requestType "github.com/Amadeus-cyf/httpServer/consts/request"
	statusCode "github.com/Amadeus-cyf/httpServer/consts/status_code"
	"github.com/Amadeus-cyf/httpServer/model"
	"github.com/Amadeus-cyf/httpServer/socketIO"
	"github.com/Amadeus-cyf/httpServer/utils"
)

var serverConfig map[string]string

var connectionPool *utils.ConnectionPool

func StartHttp() (err error) {
	if serverConfig, err = config.GetConfig(); err != nil {
		return
	}
	connectionPool = utils.NewConnectionPool(5, initConnectionToTcp)
	// defer connectionPool.Close()
	http.HandleFunc("/login", handleLoginRequest)
	http.HandleFunc("/update_avatar", handleUpdateAvatarRequest)
	http.HandleFunc("/update_nickname", handleUpdateNickNameRequest)
	http.HandleFunc("/logout", handleLogoutRequest)
	err = http.ListenAndServe(":"+serverConfig[config.WebPort], nil)
	return
}

func initConnectionToTcp() (net.Conn, error) {
	socket, err := net.DialTimeout("tcp", serverConfig[config.TcpHost]+":"+serverConfig[config.TcpPort], 5*time.Second)
	if err != nil {
		fmt.Println("Error in connection to tcp: " + err.Error())
		return nil, err
	}
	return socket, nil
}

func handleLoginRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle login request")
	var (
		err  error
		conn net.Conn
	)
	if conn, err = connectionPool.FetchConnection(); err != nil {
		fmt.Println("Connection error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	var loginParams model.LoginParams
	if err = json.NewDecoder(request.Body).Decode(&loginParams); err != nil {
		fmt.Println("handleLoginRequest, json decode error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	loginParams.RequestType = requestType.Login
	var (
		argBytes []byte
	)
	if argBytes, err = json.Marshal(loginParams); err != nil {
		fmt.Println("handleLoginRequest, json marshal error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if err = socketIO.Send(conn, argBytes); err != nil {
		fmt.Println("handleLoginRequest, send error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if err = conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		fmt.Println("handleLoginRequest, set read deadline error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	receiveResponse(conn, w)
}

func handleUpdateAvatarRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle update avatar request")
	var (
		conn net.Conn
		err  error
	)
	if conn, err = connectionPool.FetchConnection(); err != nil {
		fmt.Println("Connection error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	var (
		username []string
		ok       bool
	)
	if username, ok = request.URL.Query()["username"]; !ok {
		fmt.Println("handleUpdateAvatarRequest, username not provided")
		w.WriteHeader(statusCode.InvalidParams)
		return
	} else if len(username) == 0 {
		fmt.Println("handleUpdateAvatarRequest, empty username")
		w.WriteHeader(statusCode.InvalidParams)
		return
	}
	var cookie *http.Cookie
	if cookie, err = request.Cookie(username[0]); err != nil || cookie.Value == "" {
		fmt.Println("handleUpdateAvatarRequest, user not authorized")
		w.WriteHeader(statusCode.Unauthorized)
		return
	}
	fmt.Println("Update " + username[0] + " avatar")
	var (
		file   multipart.File
		header *multipart.FileHeader
	)
	if file, header, err = request.FormFile("data"); err != nil {
		fmt.Println("handleUpdateAvatarRequest, form file fetch error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Println("handleUpdateAvatarRequest, copy file to buffer error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	updateAvatarParams := model.NewUpdateAvatarParams(username[0], requestType.UpdateAvatar, filepath.Ext(header.Filename), buf.Bytes(), cookie.Value)
	var argBytes []byte
	if argBytes, err = json.Marshal(updateAvatarParams); err != nil {
		fmt.Println("handleUpdateAvatarRequest, json marshal error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if err = socketIO.Send(conn, argBytes); err != nil {
		fmt.Println("handleUpdateAvatarRequest, send error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if err = conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		fmt.Println("handleUpdateAvatarRequest, set read deadline error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	receiveResponse(conn, w)
}

func handleUpdateNickNameRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle update nickname request")
	var (
		conn net.Conn
		err  error
	)
	if conn, err = connectionPool.FetchConnection(); err != nil {
		fmt.Println("Connection error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	var updateNickNameParams model.UpdateNicknameParams
	if err = json.NewDecoder(request.Body).Decode(&updateNickNameParams); err != nil {
		fmt.Println("handleUpdateNicknameRequest, json decode error: ", err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if cookie, err := request.Cookie(updateNickNameParams.Username); err != nil || cookie.Value == "" {
		w.WriteHeader(statusCode.Unauthorized)
		return
	}
	updateNickNameParams.RequestType = requestType.UpdateNickname
	var argBytes []byte
	if argBytes, err = json.Marshal(updateNickNameParams); err != nil {
		fmt.Println("handleUpdateNicknameRequest, json marshal error: ", err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if err := socketIO.Send(conn, argBytes); err != nil {
		fmt.Println("handleUpdateNicknameRequest, send error: ", err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	receiveResponse(conn, w)
}

func handleLogoutRequest(w http.ResponseWriter, request *http.Request) {
	fmt.Println("handle logout request")
	var (
		conn net.Conn
		err  error
	)
	if conn, err = connectionPool.FetchConnection(); err != nil {
		fmt.Println("Connection error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	defer connectionPool.PutConnection(conn)
	var logoutParams model.LogoutParams
	if err = json.NewDecoder(request.Body).Decode(&logoutParams); err != nil {
		fmt.Println("handle logout request, json decode error: ", err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if cookie, err := request.Cookie(logoutParams.Username); err != nil || cookie.Value == "" {
		fmt.Println("user not authorized")
		w.WriteHeader(statusCode.Unauthorized)
		return
	}
	logoutParams.RequestType = requestType.Logout
	var argBytes []byte
	if argBytes, err = json.Marshal(logoutParams); err != nil {
		fmt.Println("LogoutRequest, json marshal error: ", err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	if err := socketIO.Send(conn, argBytes); err != nil {
		fmt.Println("handleUpdateNicknameRequest, send error: ", err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	receiveResponse(conn, w)
}

func receiveResponse(conn net.Conn, w http.ResponseWriter) {
	respBytes, receiveErr := socketIO.Receive(conn)
	if receiveErr != nil {
		fmt.Println("receive error: " + receiveErr.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	var resp model.HttpResponse
	var err error
	if resp, err = handleTcpResponse(respBytes); err != nil {
		fmt.Println("handle tcp response error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
		return
	}
	fmt.Println("Receive response", resp)
	if resp.RequestType == requestType.Login {
		// set cookie for http
		setCookie(w, resp)
	} else if resp.RequestType == requestType.Logout {
		deleteCookie(w, resp.Body["username"].(string))
	}
	writeResponse(w, resp)
}

func setCookie(w http.ResponseWriter, response model.HttpResponse) {
	sessionInfo := response.Body
	if sessionInfo != nil {
		cookie := http.Cookie{Name: sessionInfo["user"].(model.User).Username, Value: sessionInfo["token"].(string)}
		http.SetCookie(w, &cookie)
	}
}

func deleteCookie(w http.ResponseWriter, username string) {
	cookie := http.Cookie{Name: username, Value: "", Expires: time.Unix(0, 0)}
	http.SetCookie(w, &cookie)
}

func writeResponse(w http.ResponseWriter, response model.HttpResponse) {
	w.Header().Set("Content-type", "application/json")
	bytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("writeResponse, json marshal error: " + err.Error())
		w.WriteHeader(statusCode.ServerError)
	} else {
		w.WriteHeader(response.StatusCode)
		n, err := w.Write(bytes)
		if err != nil || n < len(bytes) {
			w.WriteHeader(statusCode.ServerError)
		}
	}
}

func handleTcpResponse(respBytes []byte) (resp model.HttpResponse, err error) {
	fmt.Println("handle tcp response")
	var response model.BaseResponse
	if err = json.Unmarshal(respBytes, &response); err != nil {
		fmt.Println("handleTcpResponse, json unmarshal error:", err.Error())
		return
	}
	switch response.RequestType {
	case requestType.Login:
		var r model.TcpLoginResponse
		if err = json.Unmarshal(respBytes, &r); err != nil {
			fmt.Println("handleTcpLoginResponse, json unmarshal error:", err.Error())
			return
		}
		resp = model.CastToHttpResponse(r)
	case requestType.UpdateAvatar, requestType.UpdateNickname:
		var r model.TcpUpdateResponse
		if err = json.Unmarshal(respBytes, &r); err != nil {
			fmt.Println("handleTcpUpdateResponse, json unmarshal error:", err.Error())
			return
		}
		resp = model.CastToHttpResponse(r)
	case requestType.Logout:
		var r model.TcpLogoutResponse
		if err = json.Unmarshal(respBytes, &r); err != nil {
			fmt.Println("handleTcpLogoutResponse, json unmarshal error", err.Error())
			return
		}
		resp = model.CastToHttpResponse(r)
	}
	return
}
