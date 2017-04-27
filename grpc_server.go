/*
 * File: grpc_server.go
 *
 * Description: contains functions to handle the initial gRPC server
 */

package main

import (
	pb "./lclusterpb"
	grpc "google.golang.org/grpc"
	"net"
)

//
// Definition of the LclusterdServer, useful with grpc proto.
//
type LclusterdServer struct {
}

//
// Jobs are merely gRPC obj refs
//
type Job pb.StartJobRequest

//! Start a grpc server instance
/*
 * @return    none
 */
func startServerInstanceOfGRPC() {

	// Listen the defined grpc port.
	listener, err := net.Listen("tcp", grpcPort)

	// Safety check, make sure an error didn't occur.
	if err != nil {
		printf(err.Error())
		panic("Error: Unable to start gRPC server on the requested port!")
	}

	// Initialize gRPC to get a server.
	remoteProcessCallServer := grpc.NewServer()

	// Mention that the grpc server has now started.
	stdlog("gRPC server startup successful.")

	// Registering grpc with the lclusterd server
	pb.RegisterLclusterdServer(remoteProcessCallServer, &LclusterdServer{})

	// Set the server to serve on the port that listener is using
	remoteProcessCallServer.Serve(listener)
}
