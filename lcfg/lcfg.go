/*
 * File: config.go
 *
 * Description: stores various hardcoded essential variables
 */

package lcfg

// If this is enabled, print debug messages.
const DebugMode = true

// Location of the etcd binary, as a POSIX dir path.
const EtcdBinaryPath = "/usr/bin/etcd"

// Storage location of the etcd server data dir; in the function
// "StartEtcdServerBackgroundProcess" found in the etcd_funcs.go file, a
// timestamp is appended to the end of this path so as to keep each server
// separate.
const EtcdDataDir = "/tmp/etcd_"

// According to the etcd documentation, the clients listen on 2379, while
// servers listen on 2380.
const EtcdClientPort = ":2379"
const EtcdServerPort = ":2380"

// IPv4 address of the gRPC server.
const GrpcServerAddr = "localhost"

// High number port for use by the gRPC server.
const GrpcPort = ":64051"

// Etcd grace period, in seconds.
const EtcdGracePeriod = 3

// Prefered command shell.
const Sh = "/bin/bash"

// Variables needed by etcd to store values.
const Nodes_dir = "/nodes"
const Processes_dir = "/processes"
const Jobs_dir = "/jobs"
const Queue_dir = "/queue"
const Failed_nodes_dir = "/failed_nodes"
const Warned_jobs_dir = "/warned_jobs"

// Location to hold which node is currently ready, aka 'primed'.
const Primed = "/primed"

// TTL lock values, in seconds.
const PrimedTTL = 4
const NlistTTL = 4

// CheckJobResponse return code constants.
//
// -1 --> failure, due to corrupted server or input
//  0 --> unknown job status
//  1 --> process does not exist
//  2 --> process is queued
//  3 --> process is active on a node
//
const CjrCorruptedServerInput = -1
const CjrUnknown = 0
const CjrProcessNotExist = 1
const CjrProcessQueued = 2
const CjrProcessActive = 3

// StopJobResponse return code constants.
//
// -1 --> failure, due to corrupted server or input
//  0 --> success
//  1 --> process does not exist
//
const SjrFailure = -1
const SjrSuccess = 0
const SjrDoesNotExist = 1

/* --------------------------------------------------
 * Capabilities the libcontainer instances will need.
 * --------------------------------------------------
 *
 * chown
 * dac_override
 * fsetid
 * fowner
 * mknod
 * net_raw
 * setgid
 * setuid
 * setfcap
 * setpcap
 * net_bind_service
 * sys_chroot
 * kill
 * audit_write
 *
 * --------------------------------------------------
 */
var LclustercCaps = []string{
	"CAP_CHOWN",
	"CAP_DAC_OVERRIDE",
	"CAP_FSETID",
	"CAP_FOWNER",
	"CAP_MKNOD",
	"CAP_NET_RAW",
	"CAP_SETGID",
	"CAP_SETUID",
	"CAP_SETFCAP",
	"CAP_SETPCAP",
	"CAP_NET_BIND_SERVICE",
	"CAP_SYS_CHROOT",
	"CAP_KILL",
	"CAP_AUDIT_WRITE",
}
