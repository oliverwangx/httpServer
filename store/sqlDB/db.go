package sqlDB

import (
	"database/sql"

	"github.com/Amadeus-cyf/httpServer/config"
	"github.com/Amadeus-cyf/httpServer/model"
)

type DBStore struct {
	db *sql.DB
}

func (d *DBStore) Init() (err error) {
	var configs map[string]string
	if configs, err = config.GetConfig(); err != nil {
		return
	}
	d.db, err = sql.Open("mysql", configs[config.SqlConn])
	return
}

func (d *DBStore) GetUserByUsername(username string, user *model.User) (err error) {
	err = d.db.QueryRow("SELECT username, password, avatar, nickname FROM User WHERE username = ?", username).Scan(user.Username, user.Password, user.Avatar, user.Nickname)
	return
}

func (d *DBStore) UpdateUserAvatar(username string, url string) (err error) {
	_, err = d.db.Exec("UPDATE User Set avatar = ? WHERE username = ?", url, username)
	return
}

func (d *DBStore) UpdateUserNickname(username string, nickname string) (err error) {
	_, err = d.db.Exec("UPDATE User Set nickname = ? WHERE username = ?", nickname, username)
	return
}

func (d *DBStore) SetUserSession(username string, token string) (err error) {
	_, err = d.db.Exec("INSERT INTO Session (Username, Token) VALUES (?, ?)", username, token)
	return
}