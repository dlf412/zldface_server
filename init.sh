#!/usr/bin/env bash

go mod init zldface_server

export GOPROXY=https://goproxy.cn
go mod tidy
echo complie...
go build -o db_migrate model/migrate/migrate.go
./db_migrate
go build -o cache_preload cache/tools/preload.go
