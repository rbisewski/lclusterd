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

	// Listen the defined grpc port.
	listener, err := net.Listen("tcp", lcfg.GrpcPort)

	// Safety check, make sure an error didn't occur.
	if err != nil {
		return err
	}

	// Initialize gRPC to get a server.
	remoteProcessCallServer := grpc.NewServer()

	// Mention that the grpc server has now started.
	stdlog("gRPC server startup successful.")

	// Registering grpc with the lclusterd server
	pb.RegisterLclusterdServer(remoteProcessCallServer, &LclusterdServer{})

	// Set the server to serve on the port that listener is using
	remoteProcessCallServer.Serve(listener)

	// Everything worked out fine, so pass nil
	return nil
}
