package limitcap

import "github.com/gin-gonic/gin"

// DefaultVisibleItems 与聊天接口、大模型 prompt 压缩层保持一致：默认最多展示 5 条。
const DefaultVisibleItems = 5

// CapGinHSlice 将列表裁剪为最多 limit 条，并返回原始总条数。
func CapGinHSlice(items []gin.H, limit int) ([]gin.H, int) {
	if limit <= 0 {
		limit = DefaultVisibleItems
	}
	total := len(items)
	if total <= limit {
		return items, total
	}
	out := make([]gin.H, limit)
	copy(out, items[:limit])
	return out, total
}

// CapAnySlice 将任意切片（[]interface{}）裁剪为最多 limit 条，并返回原始总条数。
func CapAnySlice(items []interface{}, limit int) ([]interface{}, int) {
	if limit <= 0 {
		limit = DefaultVisibleItems
	}
	total := len(items)
	if total <= limit {
		return items, total
	}
	out := make([]interface{}, limit)
	copy(out, items[:limit])
	return out, total
}
