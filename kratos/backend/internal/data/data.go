package data

import (
	"context"
	"fmt"
	"reflect"

	"github.com/bwmarrin/snowflake"
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

// sfNode snowflake 单节点生成器：应用单实例部署，node id 固定 1。
// 多实例时给每实例不同 node id（0..1023），届时从环境变量读入替换常量。
var sfNode = func() *snowflake.Node {
	n, err := snowflake.NewNode(1)
	if err != nil {
		panic(err) // 仅 node id 越界才出错，常量 1 永不触发
	}
	return n
}()

// assignSnowflakeID 在 INSERT 前为主键为 0 的记录分配 snowflake id，并关闭
// AutoIncrement 回写：列已无自增，LastInsertId 恒 0，避免 INSERT 后 gorm 把结构体
// 主键覆盖为 0（book.go/translate.go 有 Create 后立即用 b.ID/ch.ID 的调用点）。
func assignSnowflakeID(db *gorm.DB) {
	stmt := db.Statement
	f := stmt.Schema.PrioritizedPrimaryField
	if f == nil {
		return
	}
	f.AutoIncrement = false
	set := func(v reflect.Value) { // v: struct（可寻址）
		elem := reflect.Indirect(v)
		pk := elem.FieldByIndex(f.StructField.Index)
		if !pk.CanSet() {
			return // 非指针 Create 的不可寻址值：跳过，防 reflect panic
		}
		switch pk.Kind() {
		case reflect.Uint:
			if pk.Uint() == 0 {
				pk.SetUint(uint64(sfNode.Generate().Int64()))
			}
		case reflect.Int:
			if pk.Int() == 0 {
				pk.SetInt(sfNode.Generate().Int64())
			}
		}
	}
	switch stmt.ReflectValue.Kind() {
	case reflect.Struct:
		set(stmt.ReflectValue)
	case reflect.Slice:
		for i := 0; i < stmt.ReflectValue.Len(); i++ {
			set(stmt.ReflectValue.Index(i))
		}
	}
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
	// 主键改由应用层 snowflake 分配（表已去 AUTO_INCREMENT，见 sql/init.sql 迁移段）。
	// 同名字注册会覆盖而非叠加，多次 NewData 也安全。
	db.Callback().Create().Before("gorm:create").Register("snowflake:assign_id", assignSnowflakeID)
	if err := db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{mysql.Open(replica)},
	})); err != nil {
		return nil, fmt.Errorf("register replica: %w", err)
	}
	rdb := newRedis(c.RedisAddr)
	cache, err := NewCache(rdb)
	if err != nil {
		return nil, err
	}
	return &Data{DB: db, RDB: rdb, Cache: cache, ES: NewES(c.OpensearchAddr)}, nil
}

func newRedis(addr string) *redis.Client {
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

// WriteAudit 追加管理操作审计日志（best-effort，失败不阻断主流程）。
func WriteAudit(db *gorm.DB, ctx context.Context, adminID int64, action, targetType, targetID, detail string) {
	uid := adminID
	db.WithContext(ctx).Create(&AuditLog{UserID: &uid, Action: action, TargetType: targetType, TargetID: targetID, Detail: detail})
}
