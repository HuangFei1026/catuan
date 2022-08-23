package paginations

var (
	defaultPageSize    = 10
	defaultPageNo      = 1
	defaultPageSizeMax = 100
)

func SetDefaultMaxPageSize(maxSize int) {
	defaultPageSizeMax = maxSize
}

type Pagination[T any] struct {
	PageNo       int `json:"page_no"`
	PageSize     int `json:"page_size"`
	SearchOption T   `json:"search_option"`
}

func (p *Pagination[T]) Offset() int {
	if p.PageNo <= 0 {
		p.PageNo = defaultPageNo
	}
	if p.PageSize <= 0 || p.PageSize > defaultPageSizeMax {
		p.PageSize = defaultPageSize
	}
	return (p.PageNo - 1) * p.PageSize
}

func (p *Pagination[T]) Limit() int {
	return p.PageSize
}
