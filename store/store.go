package store

import (
	"context"
	"fmt"

	"github.com/Amadeus-cyf/httpServer/config"
	"github.com/Amadeus-cyf/httpServer/model"
	"github.com/Amadeus-cyf/httpServer/store/redisCache"
	"github.com/Amadeus-cyf/httpServer/store/sqlDB"
)

type DataStore struct {
	db    *sqlDB.DBStore
	cache *redisCache.CacheStore
}

func (d *DataStore) Init() (err error) {
	var serverConfig map[string]string
	if serverConfig, err = config.GetConfig(); err != nil {
		fmt.Println("Error in fetching the server config")
		return
	}
	d.cache.Init(serverConfig)
	err = d.db.Init()
	return
}

func (d *DataStore) GetUserByUsername(ctx context.Context, username string, user *model.User) (err error) {
	var ok bool
	// fetch user information from cache
	if ok, err = d.cache.GetUserByUsername(ctx, username, user); ok && err == nil {
		return
	}
	// fetch from sql database
	if err = d.db.GetUserByUsername(username, user); err != nil {
		return
	}
	// add user to cache
	err = d.cache.SetUser(ctx, username, user)
	return
}

func (d *DataStore) UpdateUserAvatar(ctx context.Context, username string, url string) (err error) {
	if err = d.db.UpdateUserAvatar(username, url); err != nil {
		return
	}
	// clear cache
	err = d.cache.ClearUser(ctx, username)
	return
}

func (d *DataStore) UpdateUserNickname(ctx context.Context, username string, nickname string) (err error) {
	if err = d.db.UpdateUserNickname(username, nickname); err != nil {
		return
	}
	err = d.cache.ClearUser(ctx, username)
	return
}

func (d *DataStore) SetUserSession(ctx context.Context, username string, token string) (err error) {
	err = d.db.SetUserSession(username, token)
	return
}
