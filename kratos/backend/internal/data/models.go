package data

import "time"

// gorm 模型（表结构与 sql/init.sql 一致，从 svc1 移植）。

type User struct {
	ID           uint64     `gorm:"column:id"`
	Username     string     `gorm:"column:username"`
	Email        string     `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	Nickname     string     `gorm:"column:nickname"`
	NicknameI18n *string    `gorm:"column:nickname_i18n;type:json"`
	Avatar       string     `gorm:"column:avatar"`
	Status       int8       `gorm:"column:status"`
	Role         int8       `gorm:"column:role"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	VipExpiresAt *time.Time `gorm:"column:vip_expires_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "novel_user" }

type AuditLog struct {
	ID         uint64    `gorm:"column:id"`
	UserID     *int64    `gorm:"column:user_id"`
	Action     string    `gorm:"column:action"`
	TargetType string    `gorm:"column:target_type"`
	TargetID   string    `gorm:"column:target_id"`
	Detail     string    `gorm:"column:detail"`
	IP         string    `gorm:"column:ip"`
	UserAgent  string    `gorm:"column:user_agent"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (AuditLog) TableName() string { return "novel_audit_log" }

type Book struct {
	ID        uint64    `gorm:"column:id"`
	Title     string    `gorm:"column:title"`
	Author    string    `gorm:"column:author"`
	Summary   string    `gorm:"column:summary"`
	Cover     string    `gorm:"column:cover"`
	Lang      string    `gorm:"column:lang"`
	IsVip     int8      `gorm:"column:is_vip"`
	Status    int8      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Book) TableName() string { return "novel_book" }

type BookTranslation struct {
	ID        uint64    `gorm:"column:id"`
	BookID    uint64    `gorm:"column:book_id"`
	Lang      string    `gorm:"column:lang"`
	Title     string    `gorm:"column:title"`
	Summary   string    `gorm:"column:summary"`
	Cover     string    `gorm:"column:cover"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (BookTranslation) TableName() string { return "novel_book_translation" }

type Category struct {
	ID        uint64    `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	ParentID  uint64    `gorm:"column:parent_id"`
	SortOrder int       `gorm:"column:sort_order"`
	Status    int8      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Category) TableName() string { return "novel_category" }

type Tag struct {
	ID        uint64    `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	Lang      string    `gorm:"column:lang"`
	Status    int8      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Tag) TableName() string { return "novel_tag" }

type BookCategory struct {
	ID         uint64 `gorm:"column:id"`
	BookID     uint64 `gorm:"column:book_id"`
	CategoryID uint64 `gorm:"column:category_id"`
}

func (BookCategory) TableName() string { return "novel_book_category" }

type BookTag struct {
	ID     uint64 `gorm:"column:id"`
	BookID uint64 `gorm:"column:book_id"`
	TagID  uint64 `gorm:"column:tag_id"`
}

func (BookTag) TableName() string { return "novel_book_tag" }

type Chapter struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	BookID    uint64    `gorm:"column:book_id"`
	ChapterNo uint32    `gorm:"column:chapter_no"`
	Title     string    `gorm:"column:title"`
	WordCount uint32    `gorm:"column:word_count"`
	IsVip     uint8     `gorm:"column:is_vip"`
	Status    uint8     `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Chapter) TableName() string { return "novel_chapter" }

type ChapterContent struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	ChapterID uint64    `gorm:"column:chapter_id"`
	Lang      string    `gorm:"column:lang"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ChapterContent) TableName() string { return "novel_chapter_content" }

type ReadingProgress struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	UserID    uint64    `gorm:"column:user_id"`
	BookID    uint64    `gorm:"column:book_id"`
	ChapterID uint64    `gorm:"column:chapter_id"`
	Position  uint32    `gorm:"column:position"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ReadingProgress) TableName() string { return "novel_reading_progress" }

type Bookshelf struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	UserID    uint64    `gorm:"column:user_id"`
	BookID    uint64    `gorm:"column:book_id"`
	SortOrder int       `gorm:"column:sort_order"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Bookshelf) TableName() string { return "novel_bookshelf" }

type Comment struct {
	ID          uint64    `gorm:"primaryKey;column:id"`
	BookID      uint64    `gorm:"column:book_id"`
	ChapterID   *uint64   `gorm:"column:chapter_id"` // NULL = 书籍级
	UserID      uint64    `gorm:"column:user_id"`
	ParentID    uint64    `gorm:"column:parent_id"`
	Content     string    `gorm:"column:content"`
	LikeCount   uint32    `gorm:"column:like_count"`
	ReportCount uint32    `gorm:"column:report_count"`
	Status      uint8     `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Comment) TableName() string { return "novel_comment" }

type Like struct {
	ID         uint64    `gorm:"primaryKey;column:id"`
	UserID     uint64    `gorm:"column:user_id"`
	TargetType uint8     `gorm:"column:target_type"`
	TargetID   uint64    `gorm:"column:target_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Like) TableName() string { return "novel_like" }

// Like 目标类型（init.sql）：1=book 2=comment 3=chapter
const (
	TargetBook    = 1
	TargetComment = 2
)

type Favorite struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	UserID    uint64    `gorm:"column:user_id"`
	BookID    uint64    `gorm:"column:book_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Favorite) TableName() string { return "novel_favorite" }

type SearchLog struct {
	ID          uint64    `gorm:"primaryKey;column:id"`
	UserID      *int64    `gorm:"column:user_id"`
	Keyword     string    `gorm:"column:keyword"`
	Lang        string    `gorm:"column:lang"`
	ResultCount uint32    `gorm:"column:result_count"`
	IP          string    `gorm:"column:ip"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (SearchLog) TableName() string { return "novel_search_log" }

type RecommendLog struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	UserID    int64     `gorm:"column:user_id"`
	BookID    uint64    `gorm:"column:book_id"`
	Strategy  string    `gorm:"column:strategy"`
	RankNo    uint32    `gorm:"column:rank_no"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (RecommendLog) TableName() string { return "novel_recommend_log" }

// ReadingLog 阅读事件日志：客户端保存进度时顺带记录，行为分析数据源。
type ReadingLog struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	UserID    uint64    `gorm:"column:user_id"`
	BookID    uint64    `gorm:"column:book_id"`
	ChapterID uint64    `gorm:"column:chapter_id"`
	Lang      string    `gorm:"column:lang"`
	Position  uint32    `gorm:"column:position"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ReadingLog) TableName() string { return "novel_reading_log" }

// 支付相关模型（T-P-01/08）。

type PaymentProvider struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	Code      string    `gorm:"column:code"`
	Lang      string    `gorm:"column:lang"`
	Region    string    `gorm:"column:region"`
	Enabled   int8      `gorm:"column:enabled"`
	Sort      int       `gorm:"column:sort"`
	Config    string    `gorm:"column:config"` // AES-GCM 加密 JSON
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (PaymentProvider) TableName() string { return "novel_payment_provider" }

type PaymentOrder struct {
	ID        uint64     `gorm:"primaryKey;column:id"`
	OrderNo   string     `gorm:"column:order_no"`
	UserID    uint64     `gorm:"column:user_id"`
	Amount    float64    `gorm:"column:amount"` // DECIMAL(10,2)；比较一律转整数分
	Currency  string     `gorm:"column:currency"`
	Provider  string     `gorm:"column:provider"`
	Status    int8       `gorm:"column:status"` // 0待支付 1已支付 2失败 3已关闭
	TxID      string     `gorm:"column:tx_id"`
	PaidAt    *time.Time `gorm:"column:paid_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (PaymentOrder) TableName() string { return "novel_payment_order" }

type VipOrder struct {
	ID        uint64     `gorm:"primaryKey;column:id"`
	OrderNo   string     `gorm:"column:order_no"`
	UserID    uint64     `gorm:"column:user_id"`
	Plan      string     `gorm:"column:plan"`
	Amount    float64    `gorm:"column:amount"`
	Currency  string     `gorm:"column:currency"`
	Status    int8       `gorm:"column:status"`
	StartAt   *time.Time `gorm:"column:start_at"`
	EndAt     *time.Time `gorm:"column:end_at"`
	PaidAt    *time.Time `gorm:"column:paid_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (VipOrder) TableName() string { return "novel_vip_order" }

// VipPlan VIP 套餐（T-P-13）：支付流程生效金额/天数的数据源，表空/缺行回退内置默认。
type VipPlan struct {
	ID          uint64    `gorm:"primaryKey;column:id"`
	PlanCode    string    `gorm:"column:plan_code"`    // monthly/quarterly/yearly
	Days        int       `gorm:"column:days"`         // 有效天数
	AmountCents int64     `gorm:"column:amount_cents"` // 金额（分）
	Currency    string    `gorm:"column:currency"`     // USD
	Label       string    `gorm:"column:label"`
	Sort        int       `gorm:"column:sort"`
	Status      int8      `gorm:"column:status"` // 0禁用 1启用
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (VipPlan) TableName() string { return "novel_vip_plan" }
