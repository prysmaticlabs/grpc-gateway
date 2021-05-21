go:
	go build -o main .
protos:
	protoc -Iproto --grpc-gateway_out=logtostderr=true,paths=source_relative:proto\
	  proto/gateway/event_source.proto
	protoc -Iproto --go_out=proto --go_opt=paths=source_relative \
    --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
    proto/gateway/event_source.proto