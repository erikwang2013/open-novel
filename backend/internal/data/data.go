package data

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/conf"
)

// Data 聚合所有持久化依赖，biz 用例层直接持有（单实现，跳过 repo 接口）。
type Data struct {
	DB    *gorm.DB
	RDB   *redis.Client
	Cache *Cache
	ES    *ES
}

func NewData(c *conf.Data) (*Data, error) {
	master := c.DbDsn
	if master == "" {
		master = "root:@tcp(127.0.0.1:3306)/novel?charset=utf8mb4&parseTime=True&loc=Local"
	}
	replica := c.DbDsnReplica
	if replica == "" {
		replica = master
	}
	db, err := gorm.Open(mysql.Open(master), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open master db: %w", err)
	}
	if err := db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{mysql.Open(replica)},
	})); err != nil {
		return nil, fmt.Errorf("register replica: %w", err)
	}
	rdb := newRedis(c.RedisAddr)
	return &Data{DB: db, RDB: rdb, Cache: NewCache(rdb), ES: NewES(c.OpensearchAddr)}, nil
}

func newRedis(addr string) *redis.Client {
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}
