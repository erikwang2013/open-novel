package service

// 管理端服务适配层：仪表盘统计 / 分类 / 标签（全部 requireAdmin，T-A-14/16）。

import (
	"context"
	"strconv"
	"time"

	adminv1 "open-novel/backend/api/admin/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type AdminService struct {
	uc *biz.AdminUsecase
	adminv1.UnimplementedAdminServer
}

func NewAdminService(uc *biz.AdminUsecase) *AdminService { return &AdminService{uc: uc} }

func (s *AdminService) GetStats(ctx context.Context, req *adminv1.GetStatsReq) (*adminv1.GetStatsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	st, err := s.uc.Stats(ctx)
	if err != nil {
		return nil, err
	}
	books := make([]*adminv1.HotBook, 0, len(st.HotBooks))
	for _, b := range st.HotBooks {
		books = append(books, &adminv1.HotBook{BookId: i64(b.BookID), Title: b.Title, Hot: b.Hot})
	}
	words := make([]*adminv1.HotKeyword, 0, len(st.HotKeywords))
	for _, w := range st.HotKeywords {
		words = append(words, &adminv1.HotKeyword{Keyword: w.Keyword, Count: w.Count})
	}
	return &adminv1.GetStatsReply{
		BookCount: st.BookCount, UserCount: st.UserCount, CommentCount: st.CommentCount, Dau: st.DAU,
		HotBooks: books, HotKeywords: words,
		OrderCount: st.OrderCount, OrderAmount: st.OrderAmount, VipCount: st.VipCount,
		TodayNewUsers: st.TodayNewUsers, PendingComments: st.PendingComments, PendingReports: st.PendingReports,
	}, nil
}

// GetReports 报表：订单收入/用户增长/VIP 订阅/内容互动（requireAdmin，日期 YYYY-MM-DD，空=近 30 天）。
func (s *AdminService) GetReports(ctx context.Context, req *adminv1.GetReportsReq) (*adminv1.GetReportsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	r, err := s.uc.Reports(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	rep := &adminv1.GetReportsReply{
		OrderReport: &adminv1.OrderReport{
			TotalCount: r.Order.TotalCount, TotalAmount: r.Order.TotalAmount,
			ByDate:    toDateAmountPoints(r.Order.ByDate),
			ByChannel: make([]*adminv1.ChannelAmountPoint, 0, len(r.Order.ByChannel)),
		},
		UserReport: &adminv1.UserReport{
			TotalUsers: r.User.TotalUsers,
			ByDate:     toDateCountPoints(r.User.ByDate),
		},
		VipReport: &adminv1.VipReport{
			TotalCount: r.Vip.TotalCount, TotalAmount: r.Vip.TotalAmount,
			ByDate: toDateAmountPoints(r.Vip.ByDate),
			ByPlan: make([]*adminv1.PlanAmountPoint, 0, len(r.Vip.ByPlan)),
		},
		ContentReport: &adminv1.ContentReport{
			BooksByDate: toDateCountPoints(r.Content.BooksByDate),
			ChaptersByDate: toDateCountPoints(r.Content.ChaptersByDate),
			CommentCount: r.Content.CommentCount, ReportCount: r.Content.ReportCount,
		},
	}
	for _, c := range r.Order.ByChannel {
		rep.OrderReport.ByChannel = append(rep.OrderReport.ByChannel, &adminv1.ChannelAmountPoint{
			Channel: c.Channel, Count: c.Count, Amount: c.Amount,
		})
	}
	for _, p := range r.Vip.ByPlan {
		rep.VipReport.ByPlan = append(rep.VipReport.ByPlan, &adminv1.PlanAmountPoint{
			PlanId: p.PlanID, PlanName: p.PlanName, Count: p.Count, Amount: p.Amount,
		})
	}
	return rep, nil
}

func toDateAmountPoints(rows []biz.DateAmountPoint) []*adminv1.DateAmountPoint {
	out := make([]*adminv1.DateAmountPoint, 0, len(rows))
	for _, p := range rows {
		out = append(out, &adminv1.DateAmountPoint{Date: p.Date, Count: p.Count, Amount: p.Amount})
	}
	return out
}

func toDateCountPoints(rows []biz.DateCountPoint) []*adminv1.DateCountPoint {
	out := make([]*adminv1.DateCountPoint, 0, len(rows))
	for _, p := range rows {
		out = append(out, &adminv1.DateCountPoint{Date: p.Date, Count: p.Count})
	}
	return out
}

// ListAuditLogs 审计日志分页查询（requireAdmin），支持用户/动作/目标/时间范围筛选。
func (s *AdminService) ListAuditLogs(ctx context.Context, req *adminv1.ListAuditLogsReq) (*adminv1.ListAuditLogsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	p := pkg.ParsePage(int32(req.Page), int32(req.PageSize))
	items, total, err := s.uc.ListAuditLogs(ctx, biz.AuditLogQuery{
		UserID: req.UserId, Action: req.Action, TargetType: req.TargetType, TargetID: req.TargetId,
		StartAt: req.StartTime, EndAt: req.EndTime,
	}, p)
	if err != nil {
		return nil, err
	}
	list := make([]*adminv1.AuditLogReply, 0, len(items))
	for _, a := range items {
		list = append(list, toAuditLogReply(&a))
	}
	return &adminv1.ListAuditLogsReply{List: list, Total: total}, nil
}

func (s *AdminService) ListCategories(ctx context.Context, req *adminv1.ListCategoriesReq) (*adminv1.ListCategoriesReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*adminv1.CategoryReply, 0, len(items))
	for _, c := range items {
		list = append(list, toCategoryReply(&c))
	}
	return &adminv1.ListCategoriesReply{List: list, Total: int64(len(list))}, nil
}

func (s *AdminService) CreateCategory(ctx context.Context, req *adminv1.CreateCategoryReq) (*adminv1.CategoryReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	c, err := s.uc.CreateCategory(ctx, req.Name, u64(req.ParentId), int(req.SortOrder))
	if err != nil {
		return nil, err
	}
	return toCategoryReply(c), nil
}

func (s *AdminService) UpdateCategory(ctx context.Context, req *adminv1.UpdateCategoryReq) (*adminv1.CategoryReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var parentID *uint64
	if req.ParentId != nil {
		v := u64(*req.ParentId)
		parentID = &v
	}
	var sortOrder *int
	if req.SortOrder != nil {
		v := int(*req.SortOrder)
		sortOrder = &v
	}
	var status *int8
	if req.Status != nil {
		v := int8(*req.Status)
		status = &v
	}
	cat, err := s.uc.UpdateCategory(ctx, c.UID, u64(req.Id), biz.CategoryUpdate{
		Name: req.Name, ParentID: parentID, SortOrder: sortOrder, Status: status,
	})
	if err != nil {
		return nil, err
	}
	return toCategoryReply(cat), nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, req *adminv1.DeleteCategoryReq) (*adminv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteCategory(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &adminv1.EmptyReply{}, nil
}

func (s *AdminService) ListTags(ctx context.Context, req *adminv1.ListTagsReq) (*adminv1.ListTagsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*adminv1.TagReply, 0, len(items))
	for _, t := range items {
		list = append(list, toTagReply(&t))
	}
	return &adminv1.ListTagsReply{List: list, Total: int64(len(list))}, nil
}

func (s *AdminService) CreateTag(ctx context.Context, req *adminv1.CreateTagReq) (*adminv1.TagReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	t, err := s.uc.CreateTag(ctx, req.Name, req.Lang)
	if err != nil {
		return nil, err
	}
	return toTagReply(t), nil
}

func (s *AdminService) UpdateTag(ctx context.Context, req *adminv1.UpdateTagReq) (*adminv1.TagReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var status *int8
	if req.Status != nil {
		v := int8(*req.Status)
		status = &v
	}
	t, err := s.uc.UpdateTag(ctx, c.UID, u64(req.Id), biz.TagUpdate{Name: req.Name, Lang: req.Lang, Status: status})
	if err != nil {
		return nil, err
	}
	return toTagReply(t), nil
}

func (s *AdminService) DeleteTag(ctx context.Context, req *adminv1.DeleteTagReq) (*adminv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteTag(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &adminv1.EmptyReply{}, nil
}

// TranslateBook 翻译书籍标题与简介（requireAdmin），源取书籍原始语言。
func (s *AdminService) TranslateBook(ctx context.Context, req *adminv1.TranslateBookReq) (*adminv1.TranslateBookReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	t, err := s.uc.TranslateBook(ctx, u64(req.BookId), req.Lang)
	if err != nil {
		return nil, err
	}
	return &adminv1.TranslateBookReply{BookId: req.BookId, Lang: req.Lang, Title: t.Title, Summary: t.Summary}, nil
}

// TranslateBookChapters 翻译书籍全部章节（requireAdmin，同步串行，单章失败不中断）。
func (s *AdminService) TranslateBookChapters(ctx context.Context, req *adminv1.TranslateBookChaptersReq) (*adminv1.TranslateBookChaptersReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	r, err := s.uc.TranslateBookChapters(ctx, u64(req.BookId), req.Lang)
	if err != nil {
		return nil, err
	}
	failed := make([]int32, len(r.FailedChapters))
	for i, no := range r.FailedChapters {
		failed[i] = int32(no)
	}
	return &adminv1.TranslateBookChaptersReply{
		BookId: req.BookId, Lang: req.Lang,
		Total: int32(r.Total), Succeeded: int32(r.Succeeded), Failed: int32(r.Failed),
		FailedChapters: failed,
	}, nil
}

func toCategoryReply(c *data.Category) *adminv1.CategoryReply {
	return &adminv1.CategoryReply{
		Id: i64(c.ID), Name: c.Name, ParentId: i64(c.ParentID),
		SortOrder: int32(c.SortOrder), Status: int32(c.Status),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func toAuditLogReply(a *data.AuditLog) *adminv1.AuditLogReply {
	uid := int64(0)
	if a.UserID != nil {
		uid = *a.UserID
	}
	tid, _ := strconv.ParseInt(a.TargetID, 10, 64) // 列存字符串，解析失败按 0
	return &adminv1.AuditLogReply{
		Id: i64(a.ID), UserId: uid, Action: a.Action, TargetType: a.TargetType,
		TargetId: tid, Detail: a.Detail, Ip: a.IP, UserAgent: a.UserAgent,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}

func toTagReply(t *data.Tag) *adminv1.TagReply {
	return &adminv1.TagReply{
		Id: i64(t.ID), Name: t.Name, Lang: t.Lang, Status: int32(t.Status),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
}
