# [Pprof](https://github.com/google/pprof)
tool for visualization and analysis of profiling data

## usage
- enable `infra/pprof` in `config.yml`
- get export of pprof: `curl http://localhost:8099/debug/pprof/goroutine --output goroutine.txt`
- visualize the pprof file: `go tool pprof -http=:8086 ./goroutine.txt`