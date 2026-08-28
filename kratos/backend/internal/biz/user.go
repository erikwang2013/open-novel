package biz

// 用户用例：注册 / 登录（JWT + refresh 轮换）/ 多语言资料（任务 #20）。
// 逻辑从 svc1 移植：bcrypt、FORCE_MASTER 防主从延迟、1062 冲突识别、审计日志。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type UserUsecase struct {
	db *gorm.DB
	am *pkg.AuthManager
}

func NewUserUsecase(d *data.Data, am *pkg.AuthManager) *UserUsecase {
	return &UserUsecase{db: d.DB, am: am}
}

type TokenResp struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         UserProfile
}

type UserProfile struct {
	ID           uint64
	Username     string
	Email        string
	Nickname     string
	NicknameI18n map[string]string
	Avatar       string
	Role         int8
	Status       int8
	CreatedAt    string
}

func (uc *UserUsecase) Register(ctx context.Context, username, password, email, nickname string) (UserProfile, error) {
	if !validUsername(username) || len(password) < 8 || len(password) > 72 {
		return UserProfile{}, pkg.ErrUserArg
	}
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		nickname = username
	}
	if len(nickname) > 64 {
		return UserProfile{}, pkg.ErrUserArg
	}
	if email != "" && !strings.Contains(email, "@") {
		return UserProfile{}, pkg.ErrUserArg
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return UserProfile{}, pkg.ErrUserInternal
	}
	u := data.User{Username: username, Email: email, PasswordHash: string(hash),
		Nickname: nickname, Status: 1, Role: 1}
	if err := uc.db.WithContext(ctx).Create(&u).Error; err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return UserProfile{}, pkg.ErrUserExists
		}
		return UserProfile{}, pkg.ErrUserInternal
	}
	return UserProfile{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Role: u.Role}, nil
}

func (uc *UserUsecase) Login(ctx context.Context, username, password string) (TokenResp, error) {
	var u data.User
	// FORCE_MASTER: 注册后可能立即登录，避免主从延迟读不到刚写入的用户
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).
		Where("username = ?", username).First(&u).Error; err != nil {
		return TokenResp{}, pkg.ErrCred
	}
	if u.Status != 1 {
		return TokenResp{}, pkg.ErrClosed
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return TokenResp{}, pkg.ErrCred
	}
	uc.db.WithContext(ctx).Model(&u).Update("last_login_at", time.Now())
	return uc.tokenResp(ctx, &u)
}

func (uc *UserUsecase) Refresh(ctx context.Context, refreshToken string) (TokenResp, error) {
	if refreshToken == "" {
		return TokenResp{}, pkg.ErrToken
	}
	uid, err := uc.am.ConsumeRefresh(ctx, refreshToken)
	if err != nil {
		return TokenResp{}, err
	}
	var u data.User
	// FORCE_MASTER: 轮换后新 token 立即生效，避免读到旧数据
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&u, uid).Error; err != nil {
		return TokenResp{}, pkg.ErrToken
	}
	if u.Status != 1 {
		return TokenResp{}, pkg.ErrClosed
	}
	return uc.tokenResp(ctx, &u)
}

// tokenResp 签发新 token 对（access 30min / refresh 7d 轮换）。
func (uc *UserUsecase) tokenResp(ctx context.Context, u *data.User) (TokenResp, error) {
	access, refresh, err := uc.am.Issue(ctx, int64(u.ID), u.Username, int32(u.Role))
	if err != nil {
		return TokenResp{}, pkg.ErrUserInternal
	}
	return TokenResp{
		AccessToken: access, RefreshToken: refresh,
		ExpiresIn: int64(uc.am.AccessTTL().Seconds()),
		User:      uc.profile(u, ""),
	}, nil
}

// WriteLoginAudit 记录登录审计（best-effort）。
func (uc *UserUsecase) WriteLoginAudit(ctx context.Context, username string, uid uint64, ip, ua string) {
	detail, _ := json.Marshal(map[string]any{"username": username})
	uid64 := int64(uid)
	uc.db.WithContext(ctx).Create(&data.AuditLog{
		UserID: &uid64, Action: "login", TargetType: "user",
		TargetID: strconv.FormatUint(uid, 10), Detail: string(detail), IP: ip, UserAgent: ua,
	})
}

type UpdateProfile struct {
	Nickname     *string
	NicknameI18n map[string]string
	Email        *string
	Avatar       *string
}

func (uc *UserUsecase) Me(ctx context.Context, uid int64) (UserProfile, error) {
	var u data.User
	// FORCE_MASTER: 更新资料后立刻读自身，避免读到旧值
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&u, uid).Error; err != nil {
		return UserProfile{}, pkg.ErrNotFound
	}
	return uc.profile(&u, ""), nil
}

func (uc *UserUsecase) UpdateMe(ctx context.Context, uid int64, req UpdateProfile, lang string) (UserProfile, error) {
	updates := map[string]any{}
	if req.Nickname != nil {
		n := strings.TrimSpace(*req.Nickname)
		if n == "" || len(n) > 64 {
			return UserProfile{}, pkg.ErrUserArg
		}
		updates["nickname"] = n
	}
	if req.NicknameI18n != nil {
		b, err := json.Marshal(req.NicknameI18n)
		if err != nil {
			return UserProfile{}, pkg.ErrUserArg
		}
		v := string(b)
		updates["nickname_i18n"] = &v
	}
	if req.Email != nil {
		if *req.Email != "" && !strings.Contains(*req.Email, "@") {
			return UserProfile{}, pkg.ErrUserArg
		}
		updates["email"] = *req.Email
	}
	if req.Avatar != nil {
		if len(*req.Avatar) > 512 {
			return UserProfile{}, pkg.ErrUserArg
		}
		updates["avatar"] = *req.Avatar
	}
	if len(updates) == 0 {
		return UserProfile{}, pkg.ErrUserArg
	}
	var u data.User
	if err := uc.db.WithContext(ctx).Model(&u).Where("id = ?", uid).Updates(updates).Error; err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return UserProfile{}, pkg.ErrUserExists
		}
		return UserProfile{}, pkg.ErrUserInternal
	}
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&u, uid).Error; err != nil {
		return UserProfile{}, pkg.ErrNotFound
	}
	return uc.profile(&u, lang), nil
}

// profile 组装资料；nickname 按 lang 从 nickname_i18n 解析，缺省回落默认昵称。
func (uc *UserUsecase) profile(u *data.User, lang string) UserProfile {
	i18n := map[string]string{}
	if u.NicknameI18n != nil && *u.NicknameI18n != "" {
		_ = json.Unmarshal([]byte(*u.NicknameI18n), &i18n)
	}
	nick := u.Nickname
	if v, ok := i18n[lang]; ok && v != "" {
		nick = v
	}
	return UserProfile{
		ID: u.ID, Username: u.Username, Email: u.Email, Nickname: nick, NicknameI18n: i18n,
		Avatar: u.Avatar, Role: u.Role, Status: u.Status,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

// ListUsers 用户列表（管理员）；search 模糊匹配 username/nickname/email。
func (uc *UserUsecase) ListUsers(ctx context.Context, search string, p pkg.Page) ([]UserProfile, int64, error) {
	q := uc.db.WithContext(ctx).Model(&data.User{})
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		q = q.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}
	var total int64
	q.Count(&total)
	var list []data.User
	q.Order("id DESC").Limit(p.PageSize).Offset(p.Offset()).Find(&list)
	items := make([]UserProfile, 0, len(list))
	for i := range list {
		items = append(items, uc.profile(&list[i], ""))
	}
	return items, total, nil
}

// SetUserStatus 封禁/解封（管理员）；status 0 封禁 1 解封，与登录校验一致（status!=1 拒绝登录）。
func (uc *UserUsecase) SetUserStatus(ctx context.Context, adminID int64, id uint64, status int8) error {
	if status != 0 && status != 1 {
		return pkg.ErrBadState
	}
	if id == 0 || uint64(adminID) == id {
		return pkg.ErrBadState // 禁止操作自己
	}
	res := uc.db.Clauses(gormdb.Write).Model(&data.User{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return pkg.ErrUserInternal
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	data.WriteAudit(uc.db, ctx, adminID, "user_status", "user", strconv.FormatUint(id, 10), fmt.Sprintf("status=%d", status))
	return nil
}

// SetUserRole 调整角色（管理员）；role 1 读者 2 作者 3 管理员。
func (uc *UserUsecase) SetUserRole(ctx context.Context, adminID int64, id uint64, role int8) error {
	if role < 1 || role > 3 {
		return pkg.ErrBadState
	}
	if id == 0 || uint64(adminID) == id {
		return pkg.ErrBadState // 禁止操作自己
	}
	res := uc.db.Clauses(gormdb.Write).Model(&data.User{}).Where("id = ?", id).Update("role", role)
	if res.Error != nil {
		return pkg.ErrUserInternal
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	data.WriteAudit(uc.db, ctx, adminID, "user_role", "user", strconv.FormatUint(id, 10), fmt.Sprintf("role=%d", role))
	return nil
}

func validUsername(s string) bool {
	if len(s) < 3 || len(s) > 32 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}
