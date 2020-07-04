## Compile proto files
Run the command below from the eventdriven directory:

### Compile proto files 

protoc --proto_path=pb pb/*.proto --go_out=plugins=grpc:pb --go_opt=paths=source_relative

### Start CockroachDB Cluster 

#### Start node 1:

cockroach start \
--insecure \
--store=ordersdb1 \
--listen-addr=localhost:26257 \
--http-addr=localhost:8080 \
--join=localhost:26257,localhost:26258,localhost:26259 \
--background

#### Start node 2:
cockroach start \
--insecure \
--store=ordersdb2 \
--listen-addr=localhost:26258 \
--http-addr=localhost:8081 \
--join=localhost:26257,localhost:26258,localhost:26259 \
--background

#### Start node 3:
cockroach start \
--insecure \
--store=ordersdb3 \
--listen-addr=localhost:26259 \
--http-addr=localhost:8082 \
--join=localhost:26257,localhost:26258,localhost:26259 \
--background
