package model

// ChunkInfo 分块信息
type ChunkInfo struct {
	Key  string // 分块对象键
	ETag string // 分块ETag（可选，用于验证）
}
