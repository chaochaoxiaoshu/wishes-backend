package utils

type Pagination struct {
	Total     int `json:"total"`
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	PageCount int `json:"pageTotal"`
}

func NewPagination(total int64, pageIndex, pageSize int) Pagination {
	pageCount := int((total + int64(pageSize) - 1) / int64(pageSize))
	return Pagination{
		Total:     int(total),
		PageIndex: pageIndex,
		PageSize:  pageSize,
		PageCount: pageCount,
	}
}
