package main

import (
	"fmt"
	"log"
	"path"
	"strconv"

	"./lcfg"
	pb "./lclusterpb"
	libetcd "./lib/etcd"
	"golang.org/x/net/context"
)

// StartJob ... the gRPC API function to start a job.
/*
 * @param     Context            given host context
 * @param     StartJobRequest    object to hold the start job request
 *
 * @return    StartJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) StartJob(ctx context.Context,
	r *pb.StartJobRequest) (*pb.StartJobResponse, error) {

	response := &pb.StartJobResponse{}

	pid, err := etcdServer.AddToGlobalQueue((libetcd.Job)(*r))
	if err != nil || pid < 0 {
		response.Pid = -1
		response.Error = err.Error()
		return response, err
	}

	response.Pid = pid
	return response, nil
}

// CheckJob ... the gRPC API function to check a job.
/*
 * @param     Context            given host context
 * @param     CheckJobRequest    object to hold the check job request
 *
 * @return    CheckJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) CheckJob(ctx context.Context,
	cjr *pb.CheckJobRequest) (*pb.CheckJobResponse, error) {

	if cjr == nil || cjr.Pid < 1 {
		return &pb.CheckJobResponse{Rc: lcfg.CjrCorruptedServerInput},
			fmt.Errorf("CheckJob() --> invalid input")
	}

	response, err := etcdServer.Client.Get(ctx, path.Join(lcfg.QueueDir,
		strconv.FormatInt(cjr.Pid, 10)))
	if err != nil || response == nil {
		return &pb.CheckJobResponse{Rc: lcfg.CjrUnknown}, err
	}

	// Cycle through all of the process refs and check if the job is active.
	for _, p := range etcdServer.ProcessesList {
		if p.Uuid == cjr.Pid {
			return &pb.CheckJobResponse{Rc: lcfg.CjrProcessActive}, nil
		}
	}

	// If the process could not be detected on the system, pass back a
	// message stating that it does not exist. Note that by the time the
	// program logic has reached this point, the job either never
	// existed or has since completed.
	//
	// TODO: suggest a feature where-by the program might one day keep
	//       track of past jobs via logging or database.
	//
	return &pb.CheckJobResponse{Rc: lcfg.CjrProcessNotExist}, nil
}

// StopJob ... the gRPC API function to stop a job.
/*
 * @param     Context            given host context
 * @param     StartJobRequest    object to hold the start job request
 *
 * @return    StartJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) StopJob(ctx context.Context,
	request *pb.StopJobRequest) (*pb.StopJobResponse, error) {

	response := &pb.StopJobResponse{}

	process, err := etcdServer.ObtainProcess(request.Pid)
	if err != nil {
		log.Println(err.Error())
		response.Rc = lcfg.SjrFailure
		response.Error = err.Error()
		return response, err
	}

	if process == nil || process.Proc == nil {
		log.Println("No such process exists with Uuid: " +
			strconv.FormatInt(request.Pid, 10))
		response.Rc = lcfg.SjrDoesNotExist
		return response, nil
	}

	// attempt to stop the given process
	err = etcdServer.StopProcess(*process)
	if err != nil {
		log.Println(err.Error())
		response.Rc = lcfg.SjrFailure
		response.Error = err.Error()
		return response, err
	}

	response.Rc = lcfg.SjrSuccess
	return response, nil
}
