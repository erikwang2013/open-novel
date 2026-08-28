package service

// 用户服务适配层：proto 消息 ↔ biz 用例。

import (
	"context"

	"open-novel/backend/api/user/v1"
	"open-novel/backend/internal/biz"
)

type UserService struct {
	uc *biz.UserUsecase
	v1.UnimplementedUserServer
}

func NewUserService(uc *biz.UserUsecase) *UserService { return &UserService{uc: uc} }

func (s *UserService) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	p, err := s.uc.Register(ctx, req.Username, req.Password, req.Email, req.Nickname)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{Id: i64(p.ID), Username: p.Username, Nickname: p.Nickname, Role: int32(p.Role)}, nil
}

func (s *UserService) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginReply, error) {
	t, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	s.uc.WriteLoginAudit(ctx, req.Username, t.User.ID, clientIP(ctx), userAgent(ctx))
	return toLoginReply(t), nil
}

func (s *UserService) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (*v1.LoginReply, error) {
	t, err := s.uc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return toLoginReply(t), nil
}

func (s *UserService) GetMe(ctx context.Context, req *v1.GetMeReq) (*v1.UserReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.uc.Me(ctx, c.UID)
	if err != nil {
		return nil, err
	}
	return toUserReply(p, pickLang(ctx, req.Lang)), nil
}

func (s *UserService) UpdateMe(ctx context.Context, req *v1.UpdateMeReq) (*v1.UserReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.uc.UpdateMe(ctx, c.UID, biz.UpdateProfile{
		Nickname:     req.Nickname,
		NicknameI18n: req.NicknameI18N,
		Email:        req.Email,
		Avatar:       req.Avatar,
	}, pickLang(ctx, req.Lang))
	if err != nil {
		return nil, err
	}
	return toUserReply(p, pickLang(ctx, req.Lang)), nil
}

func toLoginReply(t biz.TokenResp) *v1.LoginReply {
	return &v1.LoginReply{
		AccessToken: t.AccessToken, RefreshToken: t.RefreshToken,
		TokenType: "Bearer", ExpiresIn: t.ExpiresIn,
		User: toUserReply(t.User, ""),
	}
}

func toUserReply(p biz.UserProfile, lang string) *v1.UserReply {
	nick := p.Nickname
	if v, ok := p.NicknameI18n[lang]; ok && v != "" {
		nick = v
	}
	return &v1.UserReply{
		Id: i64(p.ID), Username: p.Username, Email: p.Email, Nickname: nick,
		NicknameI18N: p.NicknameI18n, Avatar: p.Avatar,
		Role: int32(p.Role), Status: int32(p.Status), CreatedAt: p.CreatedAt,
	}
}
