# zldface_server
 ## 功能
  简易的人脸识别web服务，依赖arcsoft人脸识别引擎（3.X版本）。支持人脸识别和人脸分组的功能；支持单机和多机部署

 ## 接口文档
  安装swag，执行swag init。启动服务后，访问 http://host:port/face/swagger/index.html 
 
 ## 配置
 请参考config.yaml。 虹软人脸识别引擎的appid和key请使用环境变量ARCSOFT_FACE_APPID和ARCSOFT_FACE_KEY设置 

 ## Linux快速启动
 1. 执行init.sh, 构建环境，初始化表
 2. 执行run.sh
