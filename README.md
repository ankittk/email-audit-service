# email-audit-service


1. Proto Code Generation: 

Plugins required:
- `protobuf` : `sudo apt install -y protobuf-compiler` or `brew install protobuf` (for macOS)
- `protoc-gen-go` : `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `protoc-gen-go-grpc`: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

Command to generate the Go code:
```bash
 protoc \
  --proto_path=proto \
  --go_out=proto \
  --go-grpc_out=proto \
  proto/email_audit.proto
```
