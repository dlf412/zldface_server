package request

import (
	"mime/multipart"
)

type FaceUser struct {
	Uid           string                `form:"uid" gorm:"column:uid" binding:"required"`
	FaceFeature   *multipart.FileHeader `form:"faceFeature" swaggertype:"string"`
	Name          string                `form:"name" gorm:"column:name" binding:"omitempty"`
	FaceFile      *multipart.FileHeader `form:"faceFile" binding:"omitempty,image_content_type"`
	IdFile        *multipart.FileHeader `form:"idFile" binding:"omitempty,image_content_type"`
	FaceImagePath string                `form:"faceImagePath" gorm:"column:face_image_path"`
	IdImagePath   string                `form:"idImagePath" gorm:"column:id_image_path"`
	Gid           []string              `form:"gid"`
}

type FaceUserMatch struct {
	Gid        string                `form:"gid" binding:"required_if=!OnlyUpFile"`          // 分组id
	FaceFile   *multipart.FileHeader `form:"faceFile" binding:"image_content_type,required"` // 人脸图片，建议不大于500k
	OnlyUpFile bool                  `form:"onlyUpFile"`                                     // 默认0，传1的时候则表示仅仅上传图片，不需要人脸匹配
	FilePath   string                `form:"filePath" binding:"omitempty"`
	LowScore   float32               `form:"lowScore" binding:"omitempty,min=0.7,max=1,required_with=HighScore"`
	HighScore  float32               `form:"highScore" binding:"omitempty,gtefield=LowScore,max=1.0,required_with=LowScore"`
}
