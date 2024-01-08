# Graceful Shutdown
Stopping applications might waste some resources and mess up some data. We can watch operating system signals to notify our application to handle a graceful shutdown to ensure we've handled every open connection and transaction. There are some advantages:
- Preventing data corruption: incomplete transactions.
- Avoiding resource leaks: close network connections.

## [Implementation](https://github.com/nima-abdpoor/BrainBlitz/blob/develop/internal/core/server/shutdown_hook.go)
`c := make(chan os.Signal, 1)`  
`log.Print("watching to OS signals...")`  
`signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)`  
`<-c`  
`log.Print("Recieving OS signal...")`  
`TODO: Close IO connections, complete transactions, flush logs...`
- Creating Channel to watch [OS signals](https://en.wikipedia.org/wiki/Signal_(IPC)) such as: SIGHUP, SIGINT and SIGTERM.
- Receiving OS signals and closing connections, ensuring to complete services and jobs.

### References
- [Go by Example: Signals](https://gobyexample.com/signals)
- [Mastering gRPC server with graceful shutdown within Golang’s Hexagonal Architecture](https://medium.com/@pthtantai97/mastering-grpc-server-with-graceful-shutdown-within-golangs-hexagonal-architecture-0bba657b8622)