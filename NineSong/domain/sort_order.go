package domain

type SortOrder struct {
	Sort  string `bson:"sort" json:"sort"`   // 排序字段
	Order string `bson:"order" json:"order"` // 排序方式（asc 或 desc）
}
