package redisCache

import (
	"context"
	"encoding/json"

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
	var val string
	if val, err = c.rds.Get(ctx, username).Result(); err != nil {
		return
	}
	err = json.Unmarshal([]byte(val), user)
	return
}

func (c *CacheStore) SetUser(ctx context.Context, username string, user *model.User) (err error) {
	var data []byte
	if data, err = json.Marshal(*user); err != nil {
		return
	}
	err = c.rds.Set(ctx, username, data, 0).Err()
	return
}

func (c *CacheStore) DeleteUser(ctx context.Context, username string) (err error) {
	err = c.rds.Del(ctx, username).Err()
	return
}

func (c *CacheStore) SetUserSession(ctx context.Context, username string, token string) (err error) {
	err = c.rds.Set(ctx, "Session/"+username, token, 0).Err()
	return
}

func (c *CacheStore) GetUserSession(ctx context.Context, username string) (token string, err error) {
	token, err = c.rds.Get(ctx, "Session/"+username).Result()
	return
}

func (c *CacheStore) DeleteUserSession(ctx context.Context, username string) (err error) {
	err = c.rds.Del(ctx, "Session/"+username).Err()
	return
}
