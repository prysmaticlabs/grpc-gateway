go:
	go build -o main .
protos:
	protoc -I../ -Iproto --grpc-gateway_out=logtostderr=true,paths=source_relative:proto\
	  experiment/proto/api/events.proto
	protoc -I../ -Iproto --go_out=proto --go_opt=paths=source_relative \
    --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
    experiment/proto/api/events.proto