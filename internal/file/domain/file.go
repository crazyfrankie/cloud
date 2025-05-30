package domain

type File struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	URL   string `json:"url"`
	Utime int64  `json:"utime"`
}

type Folder struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Utime int64  `json:"utime"`
	Path  string `json:"path"`
}
