#!/bin/bash

# Generates the pb2 files for go and python.

protoc --go_out=pb --go_opt=paths=source_relative \
  --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
  glucose.proto

python3 -m grpc_tools.protoc -I. \
  --python_out=py --grpc_python_out=py glucose.proto