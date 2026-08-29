package service

// CDN 厂商管理服务（管理端）：requireAdmin + proto ↔ biz 转换（镜像 payment 管理端）。

import (
	"context"
	"time"

	cdnv1 "open-novel/backend/api/cdn/v1"
	"open-novel/backend/internal/biz"
)

type CdnService struct {
	uc *biz.CdnAdminUsecase
	cdnv1.UnimplementedCdnServer
}

func NewCdnService(uc *biz.CdnAdminUsecase) *CdnService { return &CdnService{uc: uc} }

func (s *CdnService) ListProviders(ctx context.Context, req *cdnv1.ListProvidersReq) (*cdnv1.ListProvidersReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListCdnProviders(ctx)
	if err != nil {
		return nil, err
	}
	r := &cdnv1.ListProvidersReply{Total: int64(len(items))}
	for i := range items {
		r.List = append(r.List, toCdnProviderReply(&items[i]))
	}
	return r, nil
}

func (s *CdnService) CreateProvider(ctx context.Context, req *cdnv1.CreateProviderReq) (*cdnv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.CreateCdnProvider(ctx, c.UID, req.Code, int(req.Sort), req.Config)
	if err != nil {
		return nil, err
	}
	return toCdnProviderReply(it), nil
}

func (s *CdnService) UpdateProvider(ctx context.Context, req *cdnv1.UpdateProviderReq) (*cdnv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var enabled *uint8
	if req.Enabled != nil {
		v := uint8(*req.Enabled)
		enabled = &v
	}
	var sort *int32
	if req.Sort != nil {
		sort = req.Sort
	}
	it, err := s.uc.UpdateCdnProvider(ctx, c.UID, u64(req.Id), enabled, sort, req.Config)
	if err != nil {
		return nil, err
	}
	return toCdnProviderReply(it), nil
}

func (s *CdnService) DeleteProvider(ctx context.Context, req *cdnv1.DeleteProviderReq) (*cdnv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteCdnProvider(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &cdnv1.EmptyReply{}, nil
}

func (s *CdnService) ToggleProvider(ctx context.Context, req *cdnv1.ToggleProviderReq) (*cdnv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.ToggleCdnProvider(ctx, c.UID, u64(req.Id))
	if err != nil {
		return nil, err
	}
	return toCdnProviderReply(it), nil
}

// toCdnProviderReply 管理端视图：config 只出是否已配置标志，绝不回显明文。
func toCdnProviderReply(it *biz.CdnProviderItem) *cdnv1.ProviderReply {
	return &cdnv1.ProviderReply{Id: i64(it.ID), Code: it.Code, Enabled: int32(it.Enabled),
		Sort: int32(it.Sort), ConfigConfigured: it.Configured,
		CreatedAt: it.CreatedAt.Format(time.RFC3339),
		UpdatedAt: it.UpdatedAt.Format(time.RFC3339)}
}
