package model

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

func NewLoginArgs(requestType string, username string, password string) (arg LoginArgs) {
	arg = LoginArgs{
		RequestType: requestType,
		Username:    username,
		Password:    password,
	}
	return
}

func NewUpdateAvatarArgs(requestType string, username string, format string, avatar []byte) (arg UpdateAvatarArgs) {
	arg = UpdateAvatarArgs{
		RequestType: requestType,
		Username:    username,
		Format:      format,
		Avatar:      avatar,
	}
	return
}
