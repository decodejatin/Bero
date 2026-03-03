//go:generate sh -c "protoc --proto_path=../../proto --go_out=../../gen/pb --go_opt=paths=source_relative --go-grpc_out=../../gen/pb --go-grpc_opt=paths=source_relative user/user.proto job/job.proto matcher/matcher.proto"

// Package proto contains .proto source definitions for Bero's internal gRPC services.
//
// To regenerate Go stubs run from the backend/ root directory:
//
//	go generate ./proto/...
//
// Requires protoc + protoc-gen-go + protoc-gen-go-grpc:
//
//	winget install Google.Protobuf
//	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
package proto
