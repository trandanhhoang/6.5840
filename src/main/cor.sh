rm mr-*
go build -buildmode=plugin ../mrapps/wc.go
go run mrcoordinator.go pg-*.txt