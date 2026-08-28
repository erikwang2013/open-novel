package pkg

// 分页约定（规划文档 §四）：page 从 1 起，page_size 缺省 20 上限 100。

type Page struct {
	Page     int
	PageSize int
}

// ParsePage 处理 proto 查询参数；0（未传）视为缺省。
func ParsePage(page, pageSize int32) Page {
	p := Page{Page: int(page), PageSize: int(pageSize)}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	return p
}

func (p Page) Offset() int { return (p.Page - 1) * p.PageSize }
