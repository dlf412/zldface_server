package request

import (
	"mime/multipart"
)

type FaceUser struct {
	Uid           string                `form:"uid" gorm:"column:uid" binding:"required"`
	FaceFeature   *multipart.FileHeader `form:"faceFeature" swaggertype:"string"`
	Name          string                `form:"name" gorm:"column:name" binding:"omitempty"`
	FaceFile      *multipart.FileHeader `form:"faceFile" binding:"omitempty,jpg_content_type"`
	IdFile        *multipart.FileHeader `form:"idFile" binding:"omitempty,jpg_content_type"`
	FaceImagePath string                `form:"faceImagePath" gorm:"column:face_image_path"`
	Gid           []string              `form:"gid"`
}

type FaceUserMatch struct {
	Gid      string                `form:"gid"`
	FaceFile *multipart.FileHeader `form:"faceFile" binding:"required,jpg_content_type"`
}
