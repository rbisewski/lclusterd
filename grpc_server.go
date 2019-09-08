/*
 * File: grpc_server.go
 *
 * Description: contains functions to handle the initial gRPC server
 */

package main

import (
	"./lcfg"
	pb "./lclusterpb"
	grpc "google.golang.org/grpc"
	"log"
	"net"
)

// Definition of the LclusterdServer, useful with grpc proto.
type LclusterdServer struct {
}

//! Start a grpc server instance.
/*
 * @return    error    error message, if any
 */
func startGRPCServer() error {

	listener, err := net.Listen("tcp", lcfg.GrpcPort)
	if err != nil {
		return err
	}

	remoteProcessCallServer := grpc.NewServer()
	log.Println("gRPC server startup successful.")

	// Registering grpc with the lclusterd server
	pb.RegisterLclusterdServer(remoteProcessCallServer, &LclusterdServer{})
	remoteProcessCallServer.Serve(listener)
	return nil
}
