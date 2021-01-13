package request

import (
	"mime/multipart"
)

type FaceUser struct {
	Name          string                `form:"name" gorm:"column:name" binding:"omitempty"`
	Uid           string                `form:"uid" gorm:"column:uid" binding:"required"`
	FaceFile      *multipart.FileHeader `form:"faceFile" binding:"omitempty,jpg_content_type"`
	IdFile        *multipart.FileHeader `form:"idFile" binding:"omitempty,jpg_content_type"`
	FaceImagePath string                `form:"faceImagePath" gorm:"column:face_image_path"`
	FaceFeature   *multipart.FileHeader `form:"faceFeature" swaggertype:"string"`
	Gid           []string              `form:"gid"`
}

type FaceUserMatch struct {
	FaceFile *multipart.FileHeader `form:"faceFile" binding:"required,jpg_content_type"`
	Gid      string                `form:"gid"`
}
