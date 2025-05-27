package domain

type User struct {
	ID       int64  `json:"id"`
	NickName string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Birthday string `json:"birthday"`
	Ctime    int64  `json:"ctime"`
	Utime    int64  `json:"utime"`
}
