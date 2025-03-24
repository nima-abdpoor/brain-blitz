# PostgreSQL vs. MongoDB Replication

| Feature              | PostgreSQL (Streaming Replication) | MongoDB (Replica Set)                             |
|----------------------|------------------------------------|---------------------------------------------------|
| Replication Type	    | Physical (WAL logs) & Logical	     | Logical (Oplog-based)                             |
| Read-Only Replicas   | Yes (By default)	                  | Yes (Secondary nodes)                             |
| Automatic Failover   | No (Requires manual intervention)	 | Yes (Automatic election)                          |
| Data Consistency	    | Strong (With synchronous mode)	    | Eventual consistency                              |
| Partial Replication	 | Yes (Logical replication)	         | No (Full replica set)                             |
| Scalability	         | Moderate (Read replicas)	          | High (Sharding + Replica Sets)                    |
| Write Availability	  | Only primary can write	            | 	Only primary can write (except sharded clusters) |
| Performance          | Fast for OLTP workloads	           | Fast for NoSQL queries                            |
