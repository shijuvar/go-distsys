## Compile proto files
Run the command below from the gRPC directory:

protoc --proto_path=pb pb/*.proto --go_out=plugins=grpc:pb --go_opt=paths=source_relative