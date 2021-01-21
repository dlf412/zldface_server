package request

type FaceGroup struct {
	Gid  string `json:"gid" gorm:"column:gid" binding:"required"`   // 分组id
	Name string `json:"name" gorm:"column:name" binding:"required"` // 分组名
}

type FaceGroupUser struct {
	Gid  string   `json:"gid" binding:"required"`  // 分组id
	Uids []string `json:"uids" binding:"required"` // 用户id列表
}
