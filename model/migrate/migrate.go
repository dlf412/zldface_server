package main

import (
	"fmt"
	"zldface_server/config"
	"zldface_server/model"
)

func Migrate() {
	config.DB.AutoMigrate(&model.FaceGroup{})
	config.DB.AutoMigrate(&model.FaceUser{})
}

func main() {
	fmt.Println("创建数据库")
	Migrate()
}
