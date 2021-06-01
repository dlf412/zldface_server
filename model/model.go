package model

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	"zldface_server/config"
)

type G_MODEL struct {
	ID        uint           `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type FaceGroup struct {
	G_MODEL
	Gid   string     `json:"gid" gorm:"comment:group id; uniqueIndex; not null; size:20"`
	Name  string     `json:"name" gorm:"comment:group名称; not null; size:200"`
	Users []FaceUser `json:"-" gorm:"many2many:face_group_users"`
}

type FaceUser struct {
	G_MODEL
	Uid           string      `json:"uid" gorm:"comment:user id; uniqueIndex; not null; size:20"`
	FaceFeature   []byte      `json:"faceFeature,omitempty" gorm:"comment:user人脸特征; size:1032"`
	Name          string      `json:"name" gorm:"comment:user名称; not null; size:20"`
	FaceImagePath string      `json:"faceImagePath" gorm:"comment:人脸照路径; size:255"`
	IdImagePath   string      `json:"idImagePath" gorm:"comment:身份证人脸照路径; size:255"`
	Groups        []FaceGroup `json:"-" gorm:"many2many:face_group_users"`
}

func (g *FaceGroup) FaceFeatures() map[string]interface{} {

	users := []FaceUser{}
	config.DB.Model(g).Association("Users").Find(&users)

	features := map[string]interface{}{}
	for _, v := range users {
		if len(v.FaceFeature) == 1032 {
			features[v.Uid] = v.FaceFeature
		} else {
			config.Logger.Warn("人脸特征值为空或者长度非法", zap.String("uid", v.Uid), zap.Int("feature size", len(v.FaceFeature)))
		}
	}
	return features
}

func (u FaceUser) LockID() string {
	return "UL#" + u.Uid
}

func (g FaceGroup) LockID() string {
	return "GL#" + g.Gid
}
