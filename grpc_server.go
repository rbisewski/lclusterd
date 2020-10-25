package main

import (
	"log"
	"net"

	"./config"
	pb "./lclusterpb"
	grpc "google.golang.org/grpc"
)

// LclusterdServer ... main structure
type LclusterdServer struct {
}

// startGRPCServer ... start a grpc server instance.
/*
 * @return    error    error message, if any
 */
func startGRPCServer() error {

	listener, err := net.Listen("tcp", config.GrpcPort)
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
