# PostgreSQL Replication

## How does Replication work
PostgreSQL primarily uses WAL (Write-Ahead Logging) Replication, where changes are written to a log and replayed on replicas.

### Replication strategies:
#### Physical replication
> copying the byte-by-byte state of a primary database to a secondary server, which is ideal for creating an exact replica for failover scenarios, ensuring high availability.  
#### Logical replication:
> Transmits data changes at the level of individual records, offering greater flexibility regarding selective data sharing and the possibility of transformations during replication.  
> For more complex configurations — like multi-active replication.  
> If we want to replicate a single table Logical replication comes into the picture here.  

### How Logical Replication Works
Logical replication happens via streaming replication protocol.  
There will be a logical slot associated with each subscriber and subscriber specifies that while connecting to the publisher.  
Replication slot expands the flexibility of sending WAL data.  

#### Keywords
* Publisher: The primary database that sends changes (INSERT, UPDATE, DELETE) to a subscriber.
* Subscriber: The replica database that receives changes and applies them.
* Replication Slot: A mechanism to track changes for a specific subscriber.

1. **The publisher** records changes in the `pg_logical` system.
2. **Replication slot** keeps track of changes.
3. **The subscriber connects** to the publisher.
4. **Changes are streamed** to the subscriber and applied.
5. **Subscriber** remains in sync with the publisher.

