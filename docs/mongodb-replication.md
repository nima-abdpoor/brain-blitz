## MongoDB Replication Setup
This guide explains how to set up MongoDB replication using Docker containers and provides insights into the replication process, oplog management, failover handling, and best practices.

### 1. Start MongoDB with Replication Enabled
Launch the primary MongoDB node with the `--replSet` flag:
```bash
docker run -d --name BB-game-mongo1 -p 27018:27017 -v BB-game-mongo1:/data/db --network bb-network mongo:8.0 --replSet rs0
```
### 2. Start Additional MongoDB Nodes
These nodes will act as secondary replicas:
```bash
docker run -d --name BB-game-mongo2 --network bb-network mongo:8.0 --replSet rs0
docker run -d --name BB-game-mongo3 --network bb-network mongo:8.0 --replSet rs0
```
### 3. Initialize the Replica Set
Connect to the primary MongoDB container:
```bash
docker exec -it BB-game-mongo1 mongosh
```
Inside mongosh, run the following command to configure the replica set:
```javascript
rs.initiate({
  _id: "rs0",
  members: [
    { _id: 0, host: "BB-game-mongo1:27017" },
    { _id: 1, host: "BB-game-mongo2:27017" },
    { _id: 2, host: "BB-game-mongo3:27017" }
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

## Replica Set Overview
MongoDB creates a replica set configuration and assigns roles to nodes.
### 1. election process
- One of the nodes (`BB-game-mongo1`, `BB-game-mongo2`, or `BB-game-mongo3`) is elected as the Primary.
- The other two nodes become Secondary members.
- If the Primary node fails, one of the Secondaries is automatically promoted.

### 2. Replication Mechanism
- The `Primary node` handles **all write operations**.
- `Secondary nodes` replicate data from the Primary using an oplog (operations log).
- Oplog is a special collection `local.oplog.rs` that stores all write operations from the Primary.
- Secondaries continuously read from the oplog and apply changes to their own copies of the database.
### 3. Automatic Failover
- If the Primary node crashes, the other members detect it via the heartbeat mechanism (`ping` every 2 seconds).
- A new election is triggered, and one of the Secondary nodes is promoted to Primary.
- The clients automatically reconnect to the new Primary.
### 4. Read Preference Handling
- By default, clients read from the Primary.
- However, we can configure read preferences to read from Secondaries for load balancing.
### 5. Handling Network Partitions
- If a node is temporarily disconnected, it re-syncs with the Primary when it rejoins.
- MongoDB handles rollback operations if inconsistencies occur.
### 6. Verify the Replica Set Status
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
    { "_id": 0, "name": "BB-game-mongo1:27017", "stateStr": "PRIMARY" },
    { "_id": 1, "name": "BB-game-mongo2:27017", "stateStr": "SECONDARY" },
    { "_id": 2, "name": "BB-game-mongo3:27017", "stateStr": "SECONDARY" }
  ],
  "ok": 1
}
```
* **PRIMARY** → This node handles writes.
* **SECONDARY** → These nodes replicate data and handle reads if configured.
* **ARBITER** → If you have an even number of nodes, you may include an arbiter to avoid a tie during elections. Arbiters participate in elections but do not store any data.


## Oplog Management
### How to View Oplog Entries
```bash
docker exec -it BB-game-mongo1 mongosh
```
```javascript
db.getSiblingDB('local').oplog.rs.find().sort({$natural: -1}).limit(5).pretty();
```
The oplog `local.oplog.rs` is a special capped collection that records all write operations performed on the Primary node. The Secondary nodes use this log to replicate changes and keep themselves in sync with the Primary.
- Every write operation (insert, update, delete) is recorded in the oplog.
- Read operations are NOT logged in the oplog (only writes are).
- Secondary nodes continuously pull operations from the Primary’s oplog and apply them.
- MongoDB record operations in the oplog, and the Secondary nodes will replicate this insertion.
### Oplog Entry Example and Field Descriptions
```javascript
{
  "op": "i",  // Operation type (insert)
  "ns": "test.users",  // Namespace (database.collection)
  "o": { "_id": ObjectId("67deb9c494f2547fbf51e945") },  // Inserted document
  "ts": Timestamp({ "t": 1742649796, "i": 1 }),  // Timestamp
"wall": ISODate("2025-03-22T13:23:16.366Z")  // Wall-clock time
}
```
| Operation | Description                                                                             |
|-----------|-----------------------------------------------------------------------------------------|
| op        | The operation type: `i = insert`, `u = update`, `d = delete`, `c = command`, `n = noop` |
| ns        | The namespace `database.collection` where the operation occurred                        |
| o         | The document that was written to the database                                           |
| o2        | Used for updates (contains _id of the updated document)                                 |
| ts        | Timestamp of the operation (used for replication)                                       |

> "op": "n" ➡️ These are heartbeat messages to keep Secondaries in sync

### Why is Oplog Size Important?
The oplog size in MongoDB determines how long operations remain available for replication before they are overwritten.
#### What Happens If the Oplog Is Too Small?
Secondary Nodes Fall Behind (SECONDARY state → RECOVERING):
* If a Secondary node disconnects or slows down, it needs to catch up when it reconnects.
* If the required oplog entries are already overwritten, the Secondary cannot sync using the oplog.
* Full Resync Required which is expensive
#### What Happens If the Oplog Is Too Large?
* Wastes Disk Space: The oplog is stored in local and does not shrink automatically.
* Longer Startup and Recovery Time: A larger oplog means longer recovery times when a Secondary node restarts.

#### How to Check Oplog Size and Retention Time?
```javascript
db.getSiblingDB('local').oplog.rs.stats();
```
| Statistic   | Description                                                                      |
|-------------|----------------------------------------------------------------------------------|
| size        | The total size of the oplog in bytes                                             |
| storageSize | The amount of storage space the oplog uses, including any wasted space, in bytes |
| count       | The number of documents (operations) currently stored in the oplog               |
| maxSize     | The maximum size of the oplog in bytes.                                          |

#### How to Resize the Oplog?
```javascript
db.adminCommand({ replSetResizeOplog: 1, size: 5120 })
```

#### Best Practices for Oplog Configuration
> * Set the Right Size: Make sure your oplog is large enough to handle downtime without losing crucial operations. If a secondary falls behind, it will need to read older entries from the oplog to catch up
> * Monitor Oplog Lag:  Use tools like rs.printSlaveReplicationInfo() to check how far behind your secondaries are. High lag could be a sign of network or performance issues.
> * Disk I/O: Oplogs can generate a lot of I/O, especially in write-heavy workloads. Make sure your disk can handle the write operations efficiently.

#### What could go wrong?
> * Lost Oplogs: If a secondary node is offline for too long and the oplog runs out of space, it won’t be able to catch up. You’ll have to do a full data resync, which can take hours or even days with large datasets.
> * Oplog Lag: If secondaries aren’t catching up fast enough, they can fall behind the primary. This is called oplog lag, and it can impact consistency in your replica set.

#### How to monitor Replication Lag?
replication lag is the delay between the primary logging an operation in the oplog and a secondary applying it. You can monitor this lag using the command:
```javascript
rs.printSlaveReplicationInfo();
```
or
```javascript
rs.printSecondaryReplicationInfo();
```
It shows how far behind each secondary is:
```json
source: BB-game-mongo1:27017
{
  syncedTo: 'Sat Mar 22 2025 14:47:48 GMT+0000 (Coordinated Universal Time)',
  replLag: '0 secs (0 hrs) behind the primary '
}
---
source: BB-game-mongo3:27017
{
  syncedTo: 'Sat Mar 22 2025 14:47:48 GMT+0000 (Coordinated Universal Time)',
  replLag: '0 secs (0 hrs) behind the primary '
}
```

### What Happens During a Failover?
If the Primary fails, MongoDB automatically elects a new Primary. Here’s how:
#### 1️⃣ Secondary nodes detect the Primary is down using the heartbeat mechanism (ping every 2 seconds).
#### 2️⃣ An election process starts, and one of the Secondaries becomes the new Primary.
#### 3️⃣ Clients automatically reconnect to the new Primary.
#### 4️⃣ The new Primary continues accepting writes, and the old Primary (if it comes back) rolls back any uncommitted writes.

