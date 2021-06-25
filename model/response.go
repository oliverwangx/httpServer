package model

import (
	requestType "github.com/Amadeus-cyf/httpServer/consts/request"
	statusCode "github.com/Amadeus-cyf/httpServer/consts/status_code"
)

type Response interface {
	GetRequestType() string
	GetStatusCode() int
}

type BaseResponse struct {
	StatusCode  int
	RequestType string
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
	Token       string
}

type TcpUpdateResponse struct {
	StatusCode  int
	RequestType string
	User        User
}

type TcpLogoutResponse struct {
	StatusCode  int
	RequestType string
	Username    string
}

type ErrResponse struct {
	StatusCode  int
	RequestType string
}

func (r BaseResponse) GetRequestType() string {
	return r.RequestType
}

func (r BaseResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r TcpLoginResponse) GetRequestType() string {
	return r.RequestType
}

func (r TcpLoginResponse) GetStatusCode() int {
	return r.StatusCode
}

func CastToTcpLoginResp(statusCode int, requestType string, user User, token string) (resp TcpLoginResponse) {
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

func NewErrResponse() ErrResponse {
	return ErrResponse{
		RequestType: "Error",
		StatusCode:  statusCode.ServerError,
	}
}

func (r ErrResponse) GetRequestType() string {
	return "Error"
}

func (r ErrResponse) GetStatusCode() int {
	return statusCode.ServerError
}

func CastToTcpUpdateResp(statusCode int, requestType string, user User) (resp TcpUpdateResponse) {
	resp = TcpUpdateResponse{
		StatusCode:  statusCode,
		RequestType: requestType,
		User:        user,
	}
	return
}

func CastToTcpLogoutResp(statusCode int, requestType string, username string) (resp TcpLogoutResponse) {
	resp = TcpLogoutResponse{
		StatusCode:  statusCode,
		RequestType: requestType,
		Username:    username,
	}
	return
}

func (r TcpLogoutResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r TcpLogoutResponse) GetRequestType() string {
	return r.RequestType
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
		r.Body["user"] = response.(TcpLoginResponse).User
		r.Body["token"] = response.(TcpLoginResponse).Token
	case requestType.UpdateAvatar, requestType.UpdateNickname:
		r.Body["user"] = response.(TcpUpdateResponse).User
	case requestType.Logout:
		r.Body["user"] = response.(TcpLogoutResponse).Username
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
