# users-manager

## Development

```bash
go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go-grpc_out=. proto/users.proto
protoc --go_out=. --go-grpc_out=. proto/distributedTx.proto
```
## Launch

```bash
docker-compose up --build -d --remove-orphans
```

### Known problems
if api-service didn't start, just restart it
```bash
docker-compose up -d
```

### config filename

Service uses **config.yml** as a configuration file name. It can not be overriden. Probably, should be passed as a command line argument.
