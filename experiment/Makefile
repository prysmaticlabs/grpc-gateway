go:
	go build -o main .
protos:
	protoc -Iproto --grpc-gateway_out=logtostderr=true,paths=source_relative:proto\
	  proto/api/events.proto
	protoc -Iproto --go_out=proto --go_opt=paths=source_relative \
    --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
    proto/api/events.proto