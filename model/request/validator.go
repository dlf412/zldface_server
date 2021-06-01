package request

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"mime/multipart"
	"reflect"
	"strings"
	"zldface_server/config"
)

var image validator.Func = func(fl validator.FieldLevel) bool {
	im, ok := fl.Field().Interface().(*multipart.FileHeader)
	if ok {
		return im.Header.Get("Content-Type") == "image/jpeg"
	}
	return false
}

var requiredIf validator.Func = func(fl validator.FieldLevel) bool {

	/*
			Type        bool	`json:"type" validate:"required"`
			MaxValue	uint	`json:"max_value" validate:"required_if=Type"`
		    MinValue	uint	`json:"min_value" validate:"required_if=!Type"`

	*/
	var otherFieldName string
	var eq bool
	if strings.HasPrefix(fl.Param(), "!") {
		otherFieldName = fl.Param()[1:]
		eq = false
	} else {
		otherFieldName = fl.Param()
		eq = true
	}
	var otherFieldVal reflect.Value
	if fl.Parent().Kind() == reflect.Ptr {
		otherFieldVal = fl.Parent().Elem().FieldByName(otherFieldName)
	} else {
		otherFieldVal = fl.Parent().FieldByName(otherFieldName)
	}

	if isNotNilOrZeroValue(otherFieldVal) == eq {
		return isNotNilOrZeroValue(fl.Field()) // 非空
	}
	return true
}

func isNotNilOrZeroValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		return !field.IsZero()
	}
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("image_content_type", image); err != nil {
			config.Logger.Error(err.Error())
		}
		if err := v.RegisterValidation("required_if", requiredIf); err != nil {
			config.Logger.Error(err.Error())
		}
	}
}
