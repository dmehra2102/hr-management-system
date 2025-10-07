#!/bin/bash

# Generate protobuf files for HR Management System

set -e

echo "Generating protobuf files..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "protoc is not installed. Please install Protocol Buffers compiler."
    echo "Visit: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Check if Go plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "protoc-gen-go is not installed. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "protoc-gen-go-grpc is not installed. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create output directory
mkdir -p api/proto/v1/gen

# Generate Go files from proto definitions
for proto_file in api/proto/v1/*.proto; do
    echo "Generating from $proto_file..."
    protoc \
        --go_out=api/proto/v1/gen \
        --go_opt=paths=source_relative \
        --go-grpc_out=api/proto/v1/gen \
        --go-grpc_opt=paths=source_relative \
        --proto_path=api/proto/v1 \
        "$proto_file"
done

echo "✅ Protobuf generation completed successfully!"
echo "Generated files are in: api/proto/v1/gen/"
