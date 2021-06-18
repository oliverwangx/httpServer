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
}

type UpdateNicknameParams struct {
	Username    string
	RequestType string
	Nickname    string
}
