package model

type CreateFileReq struct {
	Name     string `json:"name" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
	FolderID int64  `json:"folderId"`
	URL      string `json:"url" binding:"required"`
	DeviceId string `json:"deviceId"`
}

type CreateFolderReq struct {
	Name     string `json:"name" binding:"required"`
	ParentId int64  `json:"parentId"`
}
