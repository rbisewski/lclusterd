package main

import (
	"fmt"
	"time"

	"../config"
	jobpb "../jobpb"
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
func HaveClientAddJobToServer(cmd string) (jobpb.StartJobResponse, error) {

	if len(cmd) < 1 {
		err := "Error: Given command is empty."
		return jobpb.StartJobResponse{Pid: config.SjrFailure, Error: err}, fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(config.GrpcServerAddr+config.GrpcPort,
		grpc.WithInsecure())

	if err != nil {
		return jobpb.StartJobResponse{Pid: config.SjrFailure,
			Error: err.Error()}, err
	}
	defer connection.Close()

	// Create an lcluster client
	lclusterClient := jobpb.NewLclusterdClient(connection)

	request := jobpb.StartJobRequest{
		Path:     cmd,
		Args:     []string{config.Sh, cmd},
		Env:      []string{"PATH=/bin"},
		Hostname: config.GrpcServerAddr,
	}

	// Using the new job request defined above, go ahead and start the
	// job. Retain a pointer to it so that it can be cancelled at some point.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	response, err := lclusterClient.StartJob(ctx, &request)
	cancel()

	if err != nil {
		return jobpb.StartJobResponse{}, err
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
func HaveClientCheckJobOnServer(uuid int64) (jobpb.CheckJobResponse, error) {

	if uuid < 1 {
		err := "Invalid uuid input given."
		return jobpb.CheckJobResponse{Rc: config.CjrCorruptedServerInput,
			Error: err}, fmt.Errorf(err)
	}

	// Dial a connection to the grpc server.
	connection, err := grpc.Dial(config.GrpcServerAddr+config.GrpcPort, grpc.WithInsecure())
	if err != nil {
		return jobpb.CheckJobResponse{Rc: config.CjrCorruptedServerInput,
			Error: err.Error()}, err
	}
	defer connection.Close()

	lclusterClient := jobpb.NewLclusterdClient(connection)

	request := jobpb.CheckJobRequest{Pid: uuid}

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
func HaveClientStopJobOnServer(uuid int64) (jobpb.StopJobResponse, error) {

	if uuid < 1 {
		err := "Invalid input."
		return jobpb.StopJobResponse{Rc: config.SjrFailure, Error: err},
			fmt.Errorf(err)
	}

	connection, err := grpc.Dial(config.GrpcServerAddr+config.GrpcPort, grpc.WithInsecure())
	if err != nil {
		return jobpb.StopJobResponse{Rc: config.SjrFailure,
			Error: err.Error()}, err
	}
	defer connection.Close()

	lclusterClient := jobpb.NewLclusterdClient(connection)

	request := jobpb.StopJobRequest{Pid: uuid}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	response, err := lclusterClient.StopJob(ctx, &request)
	cancel()

	return *response, err
}
