// Package proto 包含 Protobuf 定义文件和代码生成指令。
//
// 运行以下命令生成 Go 代码（从 openapi-go-sdk 根目录）：
//
//	go generate ./push/proto/
//
// 或者直接执行 protoc 命令（从 openapi-go-sdk 根目录）：
//
//	protoc --go_out=./push/pb --go_opt=paths=source_relative -I=./push/proto ./push/proto/*.proto
package proto

//go:generate protoc --go_out=../pb --go_opt=paths=source_relative -I=. ./*.proto
