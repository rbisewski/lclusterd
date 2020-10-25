package config

const (
	// DebugMode ... display debug message to stdout
	DebugMode = true

	// Rootfs path, default is ~/lclusterd/rootfs/
	Rootfs = "/lclusterd/rootfs/"

	// EtcdBinaryPath ... path to etcd
	EtcdBinaryPath = "/usr/local/bin/etcd"

	// EtcdDataDir ... storage location of the etcd server data
	EtcdDataDir = "/tmp/etcd_"

	// EtcdClientPort ... normally port 2379
	EtcdClientPort = ":2379"

	// EtcdServerPort ... normally port 2380
	EtcdServerPort = ":2380"

	// GrpcServerAddr ... IPv4 address of the gRPC server.
	GrpcServerAddr = "localhost"

	// GrpcPort ... high number port for use by the gRPC server.
	GrpcPort = ":64051"

	// EtcdGracePeriodSec ... etcd grace period, in seconds
	EtcdGracePeriodSec = 3

	// Sh ... shell
	Sh = "/bin/bash"

	// NodesDir ...
	NodesDir = "/nodes"

	// ProcessesDir ...
	ProcessesDir = "/processes"

	// JobsDir ...
	JobsDir = "/jobs"

	// QueueDir ...
	QueueDir = "/queue"

	// FailedNodesDir ...
	FailedNodesDir = "/failed_nodes"

	// WarnedJobsDir ...
	WarnedJobsDir = "/warned_jobs"

	// Primed ... location to hold which node is currently ready, aka 'primed'.
	Primed = "/primed"

	// PrimedTTL ... TTL lock values for the primed node, in seconds
	PrimedTTL = 4

	// NlistTTL ...
	NlistTTL = 4

	/* CheckJobResponse return code.
	 *
	 * -1 --> failure, due to corrupted server or input
	 *  0 --> unknown job status
	 *  1 --> process does not exist
	 *  2 --> process is queued
	 *  3 --> process is active on a node
	 */

	// CjrCorruptedServerInput ...
	CjrCorruptedServerInput = -1

	// CjrUnknown ...
	CjrUnknown = 0

	// CjrProcessNotExist ...
	CjrProcessNotExist = 1

	// CjrProcessQueued ...
	CjrProcessQueued = 2

	// CjrProcessActive ...
	CjrProcessActive = 3

	/* StopJobResponse return code.
	 *
	 * -1 --> failure, due to corrupted server or input
	 *  0 --> success
	 *  1 --> process does not exist
	 */

	// SjrFailure ... job failed to stop or was lost
	SjrFailure = -1

	// SjrSuccess ... job was stopped correctly
	SjrSuccess = 0

	// SjrDoesNotExist ... stopped job does not exist
	SjrDoesNotExist = 1
)

// LclustercCaps ... kernel capabilities needed by Lclusterc
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
