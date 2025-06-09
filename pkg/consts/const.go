package consts

const (
	DefaultAvatar = "http://localhost:9000/cloud-user/default.jpg"
	UserBucket    = "cloud-user"
	FileBucket    = "cloud-file"
)

type VIPType int

const (
	NVIP VIPType = iota
	VIP
	SVIP
)
