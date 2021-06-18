package redisCache

import (
	"context"
	"errors"
	"fmt"

	"github.com/Amadeus-cyf/httpServer/config"
	"github.com/Amadeus-cyf/httpServer/model"
	"github.com/go-redis/redis/v8"
)

type CacheStore struct {
	rds *redis.Client
}

func (c *CacheStore) Init(serverConfig map[string]string) {
	c.rds = redis.NewClient(&redis.Options{
		Addr:     serverConfig[config.RedisHost] + ":" + serverConfig[config.RedisPort],
		Password: "",
		DB:       0,
	})
}

func (c *CacheStore) GetUserByUsername(ctx context.Context, username string, user *model.User) (ok bool, err error) {
	result := c.rds.HGetAll(ctx, username)
	if result == nil {
		fmt.Println("error in get user information")
		ok, err = false, errors.New("nil result")
		return
	} else if result.Err() != nil {
		fmt.Println("redis HGet All Error: " + result.Err().Error())
		ok, err = false, result.Err()
		return
	}
	data := result.Val()
	castMapToUser(user, data)
	ok, err = true, nil
	return
}

func (c *CacheStore) SetUser(ctx context.Context, username string, user *model.User) (err error) {
	data := map[string]string{
		"username": user.Username,
		"avatar":   user.Avatar,
		"nickname": user.Nickname,
	}
	err = c.rds.HSet(ctx, username, data).Err()
	return
}

func (c *CacheStore) ClearUser(ctx context.Context, username string) (err error) {
	err = c.rds.HDel(ctx, username).Err()
	return
}

func castMapToUser(user *model.User, data map[string]string) {
	user.Username = data["username"]
	user.Avatar = data["avatar"]
	user.Nickname = data["nickname"]
}
