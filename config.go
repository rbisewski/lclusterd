/*
 * File: config.go
 *
 * Description: stores various hardcoded essential variables
 */

package main

// According to the etcd documentation, the clients listen on 2379
const etcdSocket      = "localhost:2379"

// IPv4 address of the gRPC server
const grpcServerAddr  = "localhost"

// High number port for use by the gRPC server
const grpcPort        = ":64051"

// Etcd grace period, in seconds
const etcdGracePeriod = 5
