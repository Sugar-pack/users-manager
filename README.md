# users-manager

## Development

```bash
go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go-grpc_out=. proto/users.proto
```
## Launch

```bash
docker-compose up --build -d --remove-orphans
```

* if api-service didn't start, just restart it
```bash
docker-compose up -d
```
