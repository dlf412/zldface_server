# zldface_server
 ## 简介
  基于gin框架，集成arcsoft人脸识别引擎（3.x版本），提供人脸注册、人脸匹配等功能的web api

 ## 功能特色
  1. user创建、修改、查询、user匹配查找（根据人脸）
  2. group创建、修改、查询
  3. user和group的关联关系增加和删除
  4. 人脸图片识别和匹配
  5. 支持单机和多机器部署，多机器部署依赖redis存储人脸特征

 ## 虹软人脸sdk安装说明
  下载合适的版本， 将libarcsoft_face_engine.so和libarcsoft_face.so放到 /lib/ 目录下面（linux）

 ## 接口文档
  安装swag，执行swag init。启动服务后，访问 http://host:port/face/swagger/index.html 
 
 ## 配置
 请参考config.yaml。 虹软人脸识别引擎的appid和key请使用环境变量ARCSOFT_FACE_APPID和ARCSOFT_FACE_KEY设置 

 ## Linux快速启动
 1. 执行init.sh, 构建环境，初始化表
 2. 执行run.sh
