# BrainBlitz

## 🚀 How to Run the Application

### step 1: Create shared network with docker:
```bash 
docker network create bb-network
```
### step 2: Run each service:
```bash
docker-compose -f deploy/auth/development/docker-compose.yaml up -d
docker-compose -f deploy/user/development/docker-compose.yaml up -d
docker-compose -f deploy/match/development/docker-compose.yaml up -d
docker-compose -f deploy/game/development/docker-compose.yaml up -d
```

### step 3: Run shared services in the main directory:
```bash
docker-compose up -d
```

## Replication Setup
For instructions on setting up MongoDB or PostgreSQL replication read: [MongoDB Replication Guide](docs/mongodb-replication.md), [PostgreSQL Replication Guide](docs/postgresql-replication.md)  
[PostgreSQL vs. MongoDB Replication](docs/PostgreSQL-vs-MongoDB-replication.md)

## 🤝 How to Contribute and Commit

### Protobuf
make sure to have correct package name inside your .proto file.
#### ```option go_package = "contract/[YOUR_SERVICE]/golang";```
#### Example:
option go_package = "contract/match/golang";
### How to generate .go files from .proto file
```bash
protoc --go_out=. --go-grpc_out=. contract/[YOUR_SERVICE]/proto/[YOUR_PROTO_FILE.proto]
```
#### Example:
```bash
protoc --go_out=. --go-grpc_out=. contract/match/proto/match.proto
```

Happy coding! 🚀