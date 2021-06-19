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

func (c *CacheStore) GetUserByUsername(ctx context.Context, username string) (user *model.User, err error) {
	result := c.rds.HGetAll(ctx, username)
	if result == nil {
		fmt.Println("error in get user information")
		user, err = nil, errors.New("nil result")
		return
	} else if result.Err() != nil {
		fmt.Println("redis HGet All Error: " + result.Err().Error())
		user, err = nil, result.Err()
		return
	}
	if len(result.Val()) == 0 {
		user = nil
	} else {
		user, err = castMapToUser(result.Val()), nil
	}
	return
}

func (c *CacheStore) SetUser(ctx context.Context, username string, user *model.User) (err error) {
	data := map[string]string{
		"username": user.Username,
		"avatar":   user.Avatar,
		"nickname": user.Nickname,
		"password": user.Password,
	}
	err = c.rds.HSet(ctx, username, data).Err()
	return
}

func (c *CacheStore) ClearUser(ctx context.Context, username string) (err error) {
	err = c.rds.HDel(ctx, username).Err()
	return
}

func castMapToUser(data map[string]string) (user *model.User) {
	user = &model.User{
		Username: data["username"],
		Avatar:   data["avatar"],
		Nickname: data["nickname"],
		Password: data["password"],
	}
	return
}
