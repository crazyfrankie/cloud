package model

type RegisterReq struct {
	NickName string `json:"nickname"`
	Password string `json:"password"`
}

type UpdateInfoReq struct {
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
}
