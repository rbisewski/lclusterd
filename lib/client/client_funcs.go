package libclient

import (
	"fmt"
	"time"

	"../../lcfg"
	pb "../../lclusterpb"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

//! Function so client can pass a job to the server.
/*
 * @param    string              command to execute on node
 *
 * @return   StartJobResponse    response from server
 * @return   error               error message, if any
 */
func HaveClientAddJobToServer(cmd string) (pb.StartJobResponse, error) {

	if len(cmd) < 1 {
		err := "Error: Given command is empty."
		return pb.StartJobResponse{Pid: lcfg.SjrFailure, Error: err}, fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr+lcfg.GrpcPort,
		grpc.WithInsecure())

	if err != nil {
		return pb.StartJobResponse{Pid: lcfg.SjrFailure,
			Error: err.Error()}, err
	}
	defer connection.Close()

	// Create an lcluster client
	lclusterClient := pb.NewLclusterdClient(connection)

	request := pb.StartJobRequest{
		Path:     cmd,
		Args:     []string{lcfg.Sh, cmd},
		Env:      []string{"PATH=/bin"},
		Hostname: lcfg.GrpcServerAddr,
	}

	// Using the new job request defined above, go ahead and start the
	// job. Retain a pointer to it so that it can be cancelled at some point.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	response, err := lclusterClient.StartJob(ctx, &request)
	cancel()

	if err != nil {
		return pb.StartJobResponse{}, err
	}

	return *response, nil
}

//! Function so client can check a job to the server.
/*
 * @param    string              command to execute on node
 *
 * @return   CheckJobResponse    response from server
 * @return   error               error message, if any
 */
func HaveClientCheckJobOnServer(uuid int64) (pb.CheckJobResponse, error) {

	if uuid < 1 {
		err := "Invalid uuid input given."
		return pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput,
			Error: err}, fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr+lcfg.GrpcPort, grpc.WithInsecure())
	if err != nil {
		return pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput,
			Error: err.Error()}, err
	}
	defer connection.Close()

	lclusterClient := pb.NewLclusterdClient(connection)

	request := pb.CheckJobRequest{Pid: uuid}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	response, err := lclusterClient.CheckJob(ctx, &request)
	cancel()

	return *response, err
}

//! Function so client can stop a job to the server.
/*
 * @param    int64              uuid of process to stop
 *
 * @return   StopJobResponse    response from server
 * @return   error              error message, if any
 */
func HaveClientStopJobOnServer(uuid int64) (pb.StopJobResponse, error) {

	if uuid < 1 {
		err := "Invalid input."
		return pb.StopJobResponse{Rc: lcfg.SjrFailure, Error: err},
			fmt.Errorf(err)
	}

	connection, err := grpc.Dial(lcfg.GrpcServerAddr+lcfg.GrpcPort, grpc.WithInsecure())
	if err != nil {
		return pb.StopJobResponse{Rc: lcfg.SjrFailure,
			Error: err.Error()}, err
	}
	defer connection.Close()

	lclusterClient := pb.NewLclusterdClient(connection)

	request := pb.StopJobRequest{Pid: uuid}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	response, err := lclusterClient.StopJob(ctx, &request)
	cancel()

	return *response, err
}
