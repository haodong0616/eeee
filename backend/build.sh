#!/bin/bash
# scp -r ./sqlite ./db.sqlite root@e11e:/root/go/
# scp -r ./myapp root@e11e:/root/go/
# scp -r ./sqlite ./private.pem ./public.pem ./myapp ./db.sqlite root@e11e:/root/go/
#sudo systemctl start myapp
#编译命令（只编译 main.go，排除 check_tx.go 和 test_swap.go）
CGO_ENABLED=1 CC=x86_64-linux-musl-gcc GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -extldflags "-static"' -o exchange_main main.go
scp -r ./exchange_main root@e11e:/root/go/