# MongoDB Replication Setup
This guide explains how to set up MongoDB replication using Docker containers.

### 1. Stop and Remove the Existing MongoDB Instance
If you already have a running MongoDB container, stop and remove it:
```bash
docker stop BB-game-mongo
docker rm BB-game-mongo
```
### 2. Start MongoDB with Replication Enabled
Launch the primary MongoDB node with the `--replSet` flag:
```bash
docker run -d --name BB-game-mongo -p 27018:27017 -v game-mongodb:/data/db --network bb-network mongo:8.0 --replSet rs0
```
### 3. Start Additional MongoDB Nodes
These nodes will act as secondary replicas:
```bash
docker run -d --name BB-game-mongo-2 --network bb-network mongo:8.0 --replSet rs0
docker run -d --name BB-game-mongo-3 --network bb-network mongo:8.0 --replSet rs0
```
### 4. Initialize the Replica Set
Connect to the primary MongoDB container:
```bash
docker exec -it BB-game-mongo mongosh
```
Inside mongosh, run the following command to configure the replica set:
```javascript
rs.initiate({
  _id: "rs0",
  members: [
    { _id: 0, host: "BB-game-mongo:27017" },
    { _id: 1, host: "BB-game-mongo-2:27017" },
    { _id: 2, host: "BB-game-mongo-3:27017" }
  ]
});
```
If successful, you’ll see an output similar to:
```json
{
  "ok": 1,
  "$clusterTime": { ... },
  "operationTime": { ... }
}
```
### 5. Verify the Replica Set Status
To check if all nodes are properly connected, run:
```javascript
rs.status();
```
Expected output (simplified example):
```json
{
  "set": "rs0",
  "myState": 1,
  "members": [
    { "_id": 0, "name": "BB-game-mongo:27017", "stateStr": "PRIMARY" },
    { "_id": 1, "name": "BB-game-mongo-2:27017", "stateStr": "SECONDARY" },
    { "_id": 2, "name": "BB-game-mongo-3:27017", "stateStr": "SECONDARY" }
  ],
  "ok": 1
}
```
* PRIMARY → This node handles writes.
* SECONDARY → These nodes replicate data and handle reads if configured.
### 6. Update the Go Database Connection
Modify the NewDB function to connect to the replica set:
```go
package mongo

func NewDB(config Config, ctx context.Context) (*Database, error) {
	// Use a replica set connection string
	clientOptions := options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%v,%s:%v,%s:%v/%s?replicaSet=rs0",
		config.Host, config.Port,
		"BB-game-mongo-2", 27017, // Second node
		"BB-game-mongo-3", 27017, // Third node
		config.Name,
	))

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	database := client.Database(config.Name)
	return &Database{DB: database}, nil
}
```
