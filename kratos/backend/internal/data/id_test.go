package data

// TestSnowflakeID 验证 snowflake:assign_id 回调：Create 后回读主键非零、
// 两次 Create 单调递增、批量 Create 主键非零互异。
// 需本地 MySQL(3307)/Redis(6380) 在线（与 biz 测试同款基建，见 search_test.go）。

import (
	"testing"
	"time"

	"open-novel/backend/internal/conf"
)

func TestSnowflakeID(t *testing.T) {
	d, err := NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	marker := "idtest-" + time.Now().Format("150405.000000")
	defer d.DB.Where("username LIKE ?", marker+"%").Delete(&User{})

	// 单条：主键非零 + 单调递增
	u1 := &User{Username: marker + "a", Email: marker + "a@idtest.local", Nickname: "a"}
	if err := d.DB.Create(u1).Error; err != nil {
		t.Fatal(err)
	}
	if u1.ID == 0 {
		t.Fatal("Create 后主键应为 snowflake 非零值")
	}
	u2 := &User{Username: marker + "b", Email: marker + "b@idtest.local", Nickname: "b"}
	if err := d.DB.Create(u2).Error; err != nil {
		t.Fatal(err)
	}
	if u2.ID <= u1.ID {
		t.Fatalf("snowflake id 应单调递增: %d <= %d", u2.ID, u1.ID)
	}

	// 回读一致性：库里确实存了该 id
	var got User
	if err := d.DB.First(&got, u1.ID).Error; err != nil {
		t.Fatalf("按 snowflake id 回读失败: %v", err)
	}
	if got.Username != u1.Username {
		t.Fatalf("回读 username 不一致: %q != %q", got.Username, u1.Username)
	}

	// 批量：每元素主键非零且互异
	batch := []SearchLog{{Keyword: marker + "k1"}, {Keyword: marker + "k2"}}
	if err := d.DB.Create(&batch).Error; err != nil {
		t.Fatal(err)
	}
	defer d.DB.Where("keyword LIKE ?", marker+"%").Delete(&SearchLog{})
	if batch[0].ID == 0 || batch[1].ID == 0 {
		t.Fatalf("批量 Create 主键应为非零: %d, %d", batch[0].ID, batch[1].ID)
	}
	if batch[0].ID == batch[1].ID {
		t.Fatalf("批量 Create 主键应互异: %d", batch[0].ID)
	}
}
