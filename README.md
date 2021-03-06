# lclusterd - minimalistic containers for Linux

lclusterd is a lightweight Linux job container server designed for the
purposes of running small jobs in isolated libcontainer instances.

Essentially this project can be thought of as a personal challenge, which
while technically a proof of concept, has functionality similar to docker
in roughly 2100 lines of code.

Consider reading the *Requirements* at a minimum to get the hang of how
to setup this job scheduler locally, although the author makes no promises
as to whether this will run on your specific Linux system.

To get a deeper understanding about how this program was made, read the
*Program Implementation Design* section.

Note that this is only somewhat working, as it was originally intended as a
part of a larger piece of infrastructure. It is currently uncompleted, but
occasionally it is patched to work with newer versions of libcontainer and
gRPC and golang.

# Requirements

The program itself was designed around a standard Debian Linux environment,
with the following requirements:

* linux kernel 3.19+
* golang 1.8+
* libcontainer
* docker (not actually needed, but as a quick method of assembling a rootfs)
* etcd
* grpc
* protobuf library includes
* UNIX filesystem-like structure to be mounted as a rootfs

Popular distros, such as Debian or Fedora, will likely already contain
Linux kernel 4.0 or higher. Most or all of the above requirements are
probably going to satisfied given the popularity of docker and the ease of
obtaining dependencies via go get.

In the event that this program does not appear to work on a particular
non-mainstream distro, feel free to contact me if you need assistance
and I will make note of it in future versions of this readme.

That said, an easy Makefile target has been provided to prepare a system:

```bash
make prep
```

# Installation

Having prepared your system environment as per the above section, you can
build both the server binary (lclusterd) and the client binary (lclient)
as follows:

```bash
make
```

Later on, if you need to make changes to the protobuf prototype file, you
can also run the following command:

```bash
make regen_proto
```

Note that prototypes are regenerated before every build using the standard
make target, so manual regeneration is not required.

# Basic Usage Instructions

To start the server daemon:

```bash
sudo lclusterd [--namespace=name]
```

This program will default to localhost, however, you can also specify the
network namespace using the --namespace argument flag noted above.

To add a job to the server, where 'bash command' is the terminal command
to be executed:

```bash
lclient -add='bash command'
```

To check on the current status of a job on the server, where uuid is the
assigned number of the job in question:

```bash
lclient -check=uuid
```

To remove a job from the server, use the kill argument, where uuid is
the assigned number of the job in question:

```bash
lclient -kill=uuid
```

# Program Implementation Design

The main elements of this program are:

* Scheduler, functionality to allow monitoring of a job queue.

* gRPC Server, to assist nodes and scheduler queue jobs via gRPC api calls.

* Etcd Server, which holds the values needed by this programm.

* List of nodes, with a single ready node treated as a sort of 'prime'
  node; this node is always the first node to receive jobs.

* Lcluster client, which allows end users to queue up jobs.

The goal of having a primed node is to prevent possible spamming of the
scheduler queue, since only a single node is allowed to wait for jobs as
the server is 'locked-in' to job provisioning.

After running make this program creates the following binaries:

* lclusterd --> server app
* lclient --> client app

The server requires that etcd is installed on the location mentioned in the
config.go configuration file, as per the 'EtcdBinaryPath' variable. You may
need to adjust this value to match the current location of your etcd
binary.

Eventually the goal is to have two additional elements to this software:

1) failed nodes: keep track of nodes that are inoperable during runtime.

2) warned jobs: keep track of jobs that cannot be stopped during runtime.

Another idea could be a logging or database mechanism to record previous
jobs, which would allow the end user to examine jobs ran in past days.

# Troubleshooting

If you attempt to run lclusterd and the following error occurs:

```
panic: /debug/requests is already registered. You may have two independent copies of golang.org/x/net/trace in your binary, trying to maintain separate state. This may involve a vendored copy of golang.org/x/net/trace.
```

Consider removing the golang trace routines since the CoreOS and the Etcd
imports can interfere with each other.

```bash
rm -rf $GOPATH/src/github.com/coreos/etcd/vendor/golang.org/x/net/trace
rm -rf $GOPATH/src/go.etcd.io/etcd/vendor/golang.org/x/net/trace
```

# Future Features and Todos List

Several potential ideas could perhaps be implemented to better enhance the
functionality of this program:

* merge in tests (only partly works)

* adjust namespace functionality so that it can detect network namespaces
  of a given machine

* benchmarks to ensure the program uses memory in a safe way

* handling failed nodes

* watching warned jobs

* get nodes to send console output data back to server

* consider logging / recording past jobs


# Author

This was created by Robert Bisewski at Ibis Cybernetics. For more
information, please contact us at:

* Website: www.ibiscybernetics.com

* Email: contact@ibiscybernetics.com
