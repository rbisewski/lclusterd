/*
 * File: lib/client/client_funcs.go
 *
 * Description: contains functions needed by the lclusterd client
 */

package libclient

import (
        "fmt"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"../../lcfg"
	pb "../../lclusterpb"
	"strconv"
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
	connection, err := grpc.Dial(lcfg.GrpcServerAddr + lcfg.GrpcPort,
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
	lcluster_client := pb.NewLclusterdClient(connection)

	// Start a new job request using the data obtained above.
	request := &pb.StartJobRequest{
		Path:     cmd,
		Args:     []string{lcfg.Sh, cmd},
		Env:      []string{"PATH=/bin"},
		Hostname: lcfg.GrpcServerAddr,
	}

	// Grab the current background context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Using the new job request defined above, go ahead and start the
        // job.
	response, err := lcluster_client.StartJob(ctx, request)

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
func HaveClientCheckJobOnServer(uuid string) (pb.CheckJobResponse, error) {

        // Input validation.
        if len(uuid) < 1 {
            err := "Invalid uuid input given."
            return pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput,
              Error: err}, fmt.Errorf(err)
        }

	// Attempt to cast the pid into an int since that is CheckJobRequest
	// uses as the Pid var type.
	pid, err := strconv.ParseInt(uuid, 10, 64)

	// if an error occurs, print it out
	if err != nil || pid < 1 {
            return pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput,
              Error: err.Error()}, err
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr + lcfg.GrpcPort,
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
	lcluster_client := pb.NewLclusterdClient(connection)

	// Assemble a check job request object
	request := &pb.CheckJobRequest{Pid: pid}

	// Grab the current background context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Using the check job request defined above, go ahead and attempt to
	// stop the job
	response, err := lcluster_client.CheckJob(ctx, request)

	// Cancel the current context since this has either generated a
	// response or an error.
	cancel()

        // Go ahead and return the server response and error, if any.
        return *response, err
}

//! Function so client can stop a job to the server.
/*
 * @param    string             uuid of process to stop
 *
 * @return   StopJobResponse    response from server
 * @return   error              error message, if any
 */
func HaveClientStopJobOnServer(uuid string) (pb.StopJobResponse, error) {

        // Input validation.
        if len(uuid) < 1 {
            err := "Invalid input."
            return pb.StopJobResponse{Rc: lcfg.SjrFailure, Error: err},
                   fmt.Errorf(err)
        }

	// attempt to cast the pid into an int since that is StopJobRequest
	// uses as the Pid var type
	pid, err := strconv.ParseInt(uuid, 10, 64)

	// if an error occurs, print it out
	if err != nil || pid < 1 {
            return pb.StopJobResponse{Rc: lcfg.SjrFailure,
                   Error: err.Error()}, err
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(lcfg.GrpcServerAddr + lcfg.GrpcPort,
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
	lcluster_client := pb.NewLclusterdClient(connection)

	// Assemble a stop job request object
	request := &pb.StopJobRequest{Pid: pid}

	// Grab the current background context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Using the stop job request defined above, go ahead and attempt to
	// stop the job
	response, err := lcluster_client.StopJob(ctx, request)

	// Cancel the current context since this has either generated a
	// response or an error.
	cancel()

        // Go ahead and return the response and the error, if any.
        return *response, err
}
