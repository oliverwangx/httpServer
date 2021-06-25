package model

type BaseParams struct {
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
	Token       string
}

type UpdateNicknameParams struct {
	Username    string
	RequestType string
	Nickname    string
}

type LogoutParams struct {
	Username    string
	RequestType string
}

func NewLoginParams(requestType string, username string, password string) (param LoginParams) {
	param = LoginParams{
		Username:    username,
		RequestType: requestType,
		Password:    password,
	}
	return
}

func NewUpdateAvatarParams(username string, requestType string, format string, avatar []byte, token string) (param UpdateAvatarParams) {
	param = UpdateAvatarParams{
		Username:    username,
		RequestType: requestType,
		Format:      format,
		Avatar:      avatar,
		Token:       token,
	}
	return
}
