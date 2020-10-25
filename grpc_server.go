package main

import (
	"log"
	"net"

	"./config"
	jobpb "./jobpb"
	grpc "google.golang.org/grpc"
)

// LclusterdServer ... main structure
type LclusterdServer struct {
}

// startGRPCServer ... start a grpc server instance.
func startGRPCServer() error {

	listener, err := net.Listen("tcp", config.GrpcPort)
	if err != nil {
		return err
	}

	remoteProcessCallServer := grpc.NewServer()
	log.Println("gRPC server startup successful.")

	// Registering grpc with the lclusterd server
	jobpb.RegisterLclusterdServer(remoteProcessCallServer, &LclusterdServer{})
	remoteProcessCallServer.Serve(listener)
	return nil
}
