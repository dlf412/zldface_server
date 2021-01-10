package model

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	"zldface_server/config"
)

type G_MODEL struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type FaceGroup struct {
	G_MODEL
	Gid   string     `json:"gid" gorm:"comment:group id; uniqueIndex; not null; size:50"`
	Name  string     `json:"name" gorm:"comment:group名称; not null"`
	Users []FaceUser `json:"users" gorm:"many2many:face_group_users"`
}

type FaceUser struct {
	G_MODEL
	Uid           string      `json:"uid" gorm:"comment:user id; uniqueIndex; not null; size:50"`
	Name          string      `json:"name" gorm:"comment:user名称; not null; size:20"`
	FaceFeature   []byte      `json:"faceFeature" gorm:"comment:user人脸特征; size:1032"`
	FaceImagePath string      `json:"faceImg" gorm:"comment:user人脸路径;size:255"`
	Groups        []FaceGroup `json:"groups" gorm:"many2many:face_group_users"`
}

func (g *FaceGroup) FaceFeatures() map[interface{}][]byte {

	users := []FaceUser{}
	config.DB.Model(g).Association("Users").Find(&users)

	features := map[interface{}][]byte{}
	for _, v := range users {
		if len(v.FaceFeature) == 1032 {
			features[v.Uid] = v.FaceFeature
		} else {
			config.Logger.Warn("人脸特征值为空或者长度非法", zap.String("uid", v.Uid), zap.Int("feature size", len(v.FaceFeature)))
		}
	}
	return features
}
