package model

import (
	requestType "github.com/Amadeus-cyf/httpServer/consts/request"
	statusCode "github.com/Amadeus-cyf/httpServer/consts/status_code"
)

type Response interface {
	GetRequestType() string
	GetStatusCode() int
}

type HttpResponse struct {
	StatusCode  int
	RequestType string
	Message     string
	Body        map[string]interface{}
}

type TcpLoginResponse struct {
	StatusCode  int
	RequestType string
	User        User
	Token       uint64
}

type TcpUpdateResponse struct {
	StatusCode  int
	RequestType string
	User        User
}

func (r TcpLoginResponse) GetRequestType() string {
	return r.RequestType
}

func (r TcpLoginResponse) GetStatusCode() int {
	return r.StatusCode
}

func CastToTcpLoginResp(statusCode int, requestType string, user User, token uint64) (resp TcpLoginResponse) {
	resp = TcpLoginResponse{
		StatusCode:  statusCode,
		RequestType: requestType,
		User:        user,
		Token:       token,
	}
	return
}

func (r TcpUpdateResponse) GetRequestType() string {
	return r.RequestType
}

func (r TcpUpdateResponse) GetStatusCode() int {
	return r.StatusCode
}

func CastToTcpUpdateResp(statusCode int, requestType string, user User) (resp TcpUpdateResponse) {
	resp = TcpUpdateResponse{
		StatusCode:  statusCode,
		RequestType: requestType,
		User:        user,
	}
	return
}

func CastToHttpResponse(response Response) (r HttpResponse) {
	r = HttpResponse{
		StatusCode:  response.GetStatusCode(),
		RequestType: response.GetRequestType(),
		Message:     getMessageByStatus(response.GetStatusCode()),
		Body:        make(map[string]interface{}),
	}
	switch response.GetRequestType() {
	case requestType.Login:
		r.Body["username"] = response.(TcpLoginResponse).User
		r.Body["token"] = response.(TcpLoginResponse).Token
	case requestType.UpdateAvatar:
		r.Body["username"] = response.(TcpUpdateResponse).User
	case requestType.UpdateNickname:
		r.Body["username"] = response.(TcpUpdateResponse).User
	}
	return
}

func getMessageByStatus(status int) (m string) {
	switch status {
	case statusCode.Success:
		m = "Success"
	case statusCode.NotFound:
		m = "Not Found"
	case statusCode.BadRequest:
		m = "Bad Request"
	case statusCode.Unauthorized:
		m = "User is not authorized"
	case statusCode.ServerError:
		m = "Server Error"
	case statusCode.InvalidParams:
		m = "Invalid params"
	default:
		m = ""
	}
	return
}
