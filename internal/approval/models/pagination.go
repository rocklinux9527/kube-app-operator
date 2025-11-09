package models

// Pagination 通用分页结果
type Pagination struct {
    Page      int         `json:"page"`
    Limit     int         `json:"limit"`
    Total     int64       `json:"total"`
    Data      interface{} `json:"data"`
}

