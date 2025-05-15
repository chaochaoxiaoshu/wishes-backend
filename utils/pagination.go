package utils

// Pagination 分页信息结构体
type Pagination struct {
	Total     int `json:"total"`
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	PageCount int `json:"pageTotal"`
}

// 可以添加一个辅助函数来创建分页对象
func NewPagination(total int64, pageIndex, pageSize int) Pagination {
	pageCount := int((total + int64(pageSize) - 1) / int64(pageSize))
	return Pagination{
		Total:     int(total),
		PageIndex: pageIndex,
		PageSize:  pageSize,
		PageCount: pageCount,
	}
}
