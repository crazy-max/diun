package pb

//go:generate protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. image.proto notif.proto healthcheck.proto
