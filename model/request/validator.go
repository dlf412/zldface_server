package request

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"mime/multipart"
	"zldface_server/config"
)

func imageContentType(fl validator.FieldLevel) bool {
	im, ok := fl.Field().Interface().(*multipart.FileHeader)
	if ok {
		return im.Header.Get("Content-Type") == "image/jpeg"
	}
	return false
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("jpg_content_type", imageContentType); err != nil {
			config.Logger.Error(err.Error())
		}
	}
}
