package model

type LoginReq struct {
	NickName string `json:"nickname"`
	Password string `yaml:"password"`
}
