/*
 * File: lib/client/client_funcs.go
 *
 * Description: contains functions needed by the lclusterd client
 */

package libclient

import (
	"../../lcfg"
	pb "../../lclusterpb"
	"fmt"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"time"
)

//! Function so client can pass a job to the server.
/*
 * @param    string              command to execute on node
 *
 * @return   StartJobResponse    response from server
 * @return   error               error message, if any
 */
func HaveClientAddJobToServer(cmd string) (pb.StartJobResponse, error) {

	// Input validation.
	if len(cmd) < 1 {
		err := "Error: Given command is empty."
		return pb.StartJobResponse{Pid: lcfg.SjrFailure, Error: err},
			fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr+lcfg.GrpcPort,
		grpc.WithInsecure())

	// Safety check, ensure no errors have occurred.
	if err != nil {
		return pb.StartJobResponse{Pid: lcfg.SjrFailure,
			Error: err.Error()}, err
	}

	// Defer the connection for the time being, but eventually it will be
	// closed once we've finished with it.
	defer connection.Close()

	// Create an lcluster client
	lclusterClient := pb.NewLclusterdClient(connection)

	// Start a new job request using the data obtained above.
	request := pb.StartJobRequest{
		Path:     cmd,
		Args:     []string{lcfg.Sh, cmd},
		Env:      []string{"PATH=/bin"},
		Hostname: lcfg.GrpcServerAddr,
	}

	// Grab the current background context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Using the new job request defined above, go ahead and start the
	// job.
	response, err := lclusterClient.StartJob(ctx, &request)

	// Cancel the current context since this has either generated a
	// response or an error.
	cancel()

	// return the server response and the error
	return *response, err
}

//! Function so client can check a job to the server.
/*
 * @param    string              command to execute on node
 *
 * @return   CheckJobResponse    response from server
 * @return   error               error message, if any
 */
func HaveClientCheckJobOnServer(uuid int64) (pb.CheckJobResponse, error) {

	// Input validation.
	if uuid < 1 {
		err := "Invalid uuid input given."
		return pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput,
			Error: err}, fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr+lcfg.GrpcPort,
		grpc.WithInsecure())

	// Safety check, ensure no errors have occurred.
	if err != nil {
		return pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput,
			Error: err.Error()}, err
	}

	// Defer the connection for the time being, but eventually it will be
	// closed once we've finished with it.
	defer connection.Close()

	// Create an lcluster client
	lclusterClient := pb.NewLclusterdClient(connection)

	// Assemble a check job request object
	request := pb.CheckJobRequest{Pid: uuid}

	// Grab the current background context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Using the check job request defined above, go ahead and attempt to
	// stop the job
	response, err := lclusterClient.CheckJob(ctx, &request)

	// Cancel the current context since this has either generated a
	// response or an error.
	cancel()

	// Go ahead and return the server response and error, if any.
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

	// Input validation.
	if uuid < 1 {
		err := "Invalid input."
		return pb.StopJobResponse{Rc: lcfg.SjrFailure, Error: err},
			fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr+lcfg.GrpcPort,
		grpc.WithInsecure())

	// Safety check, ensure no errors have occurred.
	if err != nil {
		return pb.StopJobResponse{Rc: lcfg.SjrFailure,
			Error: err.Error()}, err
	}

	// Defer the connection for the time being, but eventually it will be
	// closed once we've finished with it.
	defer connection.Close()

	// Create an lcluster client
	lclusterClient := pb.NewLclusterdClient(connection)

	// Assemble a stop job request object
	request := pb.StopJobRequest{Pid: uuid}

	// Grab the current background context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Using the stop job request defined above, go ahead and attempt to
	// stop the job
	response, err := lclusterClient.StopJob(ctx, &request)

	// Cancel the current context since this has either generated a
	// response or an error.
	cancel()

	// Go ahead and return the response and the error, if any.
	return *response, err
}
