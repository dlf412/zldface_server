package request

import "mime/multipart"

type FaceFeatures struct {
	Feature1 string `form:"feature1" json:"feature1" binding:"required,len=1376,base64"`
	Feature2 string `form:"feature2" json:"feature2" binding:"required,len=1376,base64"`
}

type FaceFiles struct {
	FaceFile1 *multipart.FileHeader `form:"faceFile1" binding:"required_without=Face1"` // 人脸图片文件1
	FaceFile2 *multipart.FileHeader `form:"faceFile2" binding:"required_without=Face2"` // 人脸图片文件2
	Face1     string                `form:"face1" binding:"required_without=FaceFile1"` // 人脸特征或路径
	Face2     string                `form:"face2" binding:"required_without=FaceFile2"` // 人脸特征或路径
}
