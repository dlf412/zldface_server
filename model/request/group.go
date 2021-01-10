package request

type FaceGroup struct {
	Name string `json:"name" gorm:"column:name" binding:"required"`
	Gid  string `json:"gid" gorm:"column:gid" binding:"required"`
}

type FaceGroupUser struct {
	Gid  string   `json:"gid" binding:"required"`
	Uids []string `json:"uids" binding:"required"`
}
