set MONGODB_URI=mongodb://root:password@127.0.0.1:27017/admin
rem go run main.go  --log.level=info --collector.diagnosticdata --collector.dbstats --collector.topmetrics --collector.replicasetstatus --discovering-mode --compatible-mode

go run main.go  --log.level=debug --collector.profile