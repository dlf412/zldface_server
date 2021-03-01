#!/usr/bin/env bash

if [ ! -f "go.mod" ];then
  export GOPROXY=https://goproxy.cn
  go mod init zldface_server
fi
echo complie...
# 替换docs的Host
domain=${MAIN_DOMAIN#*//}
domain=${domain%/*}
if [ $domain ]; then
  sed -i "s/localhost:8888/$domain/g" docs/*
fi
go build -o zldface_server main.go
./zldface_server