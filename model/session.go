package model

type Session struct {
	Username string
	Token    uint64
}

func NewSession(username string, token uint64) (s Session) {
	s = Session{
		Username: username,
		Token:    token,
	}
	return
}
