/*
 * File: config.go
 *
 * Description: stores various hardcoded essential variables
 */

package main

// location of the etcd binary, as a POSIX dir path
const etcdBinaryPath   = "/usr/bin/etcd"

// Storage location of the etcd server data dir; in the function
// "StartEtcdServerBackgroundProcess" found in the etcd_funcs.go file, a
// timestamp is appended to the end of this path so as to keep each server
// separate.
const etcdDataDir      = "/tmp/etcd_"

// According to the etcd documentation, the clients listen on 2379, while
// servers listen on 2380.
const etcdClientPort   = ":2379"
const etcdServerPort   = ":2380"

// IPv4 address of the gRPC server
const grpcServerAddr   = "localhost"

// High number port for use by the gRPC server
const grpcPort         = ":64051"

// Etcd grace period, in seconds
const etcdGracePeriod  = 3

// Prefered Command Shell 
const sh               = "/bin/bash"

// Variables needed by etcd to store values.
const nodes_dir        = "/nodes"
const processes_dir    = "/processes"
const jobs_dir         = "/jobs"
const queue_dir        = "/queue"
const failed_nodes_dir = "/failed_nodes"
const warned_jobs_dir  = "/warned_jobs"

// Location to hold which node is currently ready, aka 'primed'
const primed           = "/primed"

// TTL lock values, in seconds
const primedTTL        = 4
const nlistTTL         = 4

/* --------------------------------------------------
 * Capabilities the libcontainer instances will need:
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
var lclusterc_caps = []string {
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
