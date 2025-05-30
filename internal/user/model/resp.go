package model

type UserResp struct {
	ID       int64  `json:"id"`
	NickName string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Birthday string `json:"birthday"`
	Utime    int64  `json:"utime"`
}
