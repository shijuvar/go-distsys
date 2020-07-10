## Compile proto files
Run the command below from the gRPC directory:

protoc --proto_path=pb pb/*.proto --go_out=plugins=grpc:pb --go_opt=paths=source_relative

// V2

protoc --proto_path=pbv2 pbv2/*.proto --go_out=plugins=grpc:pbv2 --go_opt=paths=source_relative