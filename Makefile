#
# Makefile for lclusterd
#

# 
# phony targets
#
.PHONY: all client regen_proto clean 

#
# make targets
#

all: build

protobuf:
	@wget "https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip" -qO /tmp/protoc.zip
	@mkdir -p /tmp/protoc
	@unzip /tmp/protoc.zip -d /tmp/protoc
	@rm -f /tmp/protoc.zip
	@sudo rm -rf /usr/local/include/google/protobuf
	@sudo mkdir -p /usr/local/include/google/
	@sudo mv /tmp/protoc/include/google/protobuf /usr/local/include/google/protobuf
	@sudo mv /tmp/protoc/bin/protoc /usr/local/bin/protoc
	@rm -rf /tmp/protoc

grpc_deps:
	@echo "Obtaining gRPC..."
	@go get -u google.golang.org/grpc
	@go get -u github.com/golang/protobuf/protoc-gen-go

etcd_docker:
	@echo "Setting up etcd via docker..."
	@docker run -d --name etcd-server \
	      --publish 2379:2379 \
	      --publish 2380:2380 \
	      --env ALLOW_NONE_AUTHENTICATION=yes \
	      --env ETCD_ADVERTISE_CLIENT_URLS=http://etcd-server:2379 \
	      bitnami/etcd:latest

example_rootfs:
	@go get github.com/opencontainers/runc/libcontainer
	@docker pull ubuntu:latest
	@docker create --name ubuntu-husk ubuntu:latest
	@mkdir -p ~/lclusterd/rootfs
	@docker export ubuntu-husk | tar xvfC - ~/lclusterd/rootfs
	@docker rm ubuntu-husk

regen_proto:
	@echo "Regenerating lclusterpb protobuffs..."
	@protoc -I lclusterpb/ lclusterpb/lclusterpb.proto --go_out=plugins=grpc:lclusterpb

lclusterd:
	@echo "Building lclusterd..."
	@go build

client:
	@echo "Building lcluster client..."
	@go build -o lclusterc ./client

prep: grpc_deps etcd_docker example_rootfs
	@echo "Setting up the infrastructure needed for lclusterd..."

build: regen_proto lclusterd client 

clean:
	@echo "Cleaning binaries..."
	@go clean
	@rm -f lclusterc
	@echo "Removing possible configs generated by clients..."
	@echo "Removing generated protobuff code..."
	@rm -f lclusterpb/lclusterpb.pb.go
