GO111MODULE=on
PATH := "${PATH}:$(shell go env GOPATH)/bin"

export GO111MODULE
export PATH
export CGO_ENABLED=0

default: user.pb user.pb.gw  compile

user.pb: 
	protoc -I . -I /usr/local/include \
		--go_out=paths=source_relative:.   \
		./proto/user.proto

user_grpc.pb: 
	protoc -I . -I /usr/local/include \
		--go-grpc_out=paths=source_relative:.   \
		./proto/user.proto

user.pb.gw: 
	protoc -I . -I /usr/local/include \
		--grpc-gateway_out . \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
		./proto/user.proto

prepare:
	go get google.golang.org/grpc
	go get github.com/golang/protobuf/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

compile:
	go build -o bin/server server/main.go 
	go build -o bin/gateway grpc-gw-server/gateway.go
	go build -o bin/client client/main.go
