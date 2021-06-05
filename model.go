package main

type Response struct {
	StatusCode int
	Request    string
	Message    string
	Body       interface{}
}

type Session struct {
	Username string
	Token    string
}

type CoreParams struct {
	Username    string
	RequestType string
}

type LoginParams struct {
	Username    string
	RequestType string
	Password    string
}

type UpdateAvatarParams struct {
	Username    string
	RequestType string
	Format      string
	Avatar      []byte
}

type UpdateNicknameParams struct {
	Username    string
	RequestType string
	Nickname    string
}

type LoginArgs struct {
	RequestType string
	Username    string
	Password    string
}

type UpdateAvatarArgs struct {
	RequestType string
	Username    string
	Format      string
	Avatar      []byte
}

type UpdateNicknameArgs struct {
	RequestType string
	Username    string
	Nickname    string
}

type LogoutArgs struct {
	Username string
}
