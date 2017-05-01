# lclusterd - a lite clustered job scheduler for Linux

lcluster is a lightweight Linux job scheduler designed to create clustered
nodes for the purposes of running small jobs in isolated libcontainer
instances.

Consider reading the *Basic Usage Instructions* at a minimum to get the hang
of using this job scheduler.

Note that while this is a work in progress, it does somewhat partially work.

# Requirements

The program itself was designed around a standard Debian Linux environment,
with the following requirements:

* docker
* etcd 
* libcontainer 
* linux kernel 3.19+
* golang 1.8+
* grpc

Popular distros, such as Debian or Fedora, will likely already contain
Linux kernel 4.0 or higher. Most or all of the above requirements are
probably going to satisfied given the popularity of docker and the ease of
obtaining dependencies via go get.

In the event that this program does not appear to work on a particular
non-mainstream distro. Feel free to contact me if you need assistance
and I will make note of it in future versions of this readme.


# Installation

You can obtain the codebase via git:

    git clone https://github.com/rbisewski/lclusterd

Afterwards you can build both the server binary (lclusterd) and the client
binary (lclusterc) as follows:

    make

If you need to make changes to the PB prototype file, you can also run the
following command:

    make regen_proto

Note that prototypes are regenerated before every build using the standard
make target, so manual regeneration is not required.


# Basic Usage Instructions

Etcd is required for holding the key value stores, you can run it as
follows:

    etcd --data-dir /tmp/etcd


This program uses runc and as a result needs a rootfs to work. The fastest
way would be to export a docker image to /tmp/ and pass along the final
location to the lclusterd server. 

    mkdir /tmp/rootfs

    sudo docker export 719ae7c313c9 | tar xvfC - /tmp/rootfs


To start the server daemon:

    sudo lclusterd [--namespace=name] [--rootfs=/path/to/dir]

This program will default to localhost, however, you can also specify the
network namespace using the --namespace argument flag noted above.

Since this program uses libcontainer, you will need to specify a rootfs
location as well, which can be any safe POSIX location.


To add a job to the server, you can use the addjob argument like so:

    lclusterc --addjob='bash command'

Where 'bash command' is the terminal command to be executed.


To check on the current status of a job on the server, you can use the
checkjob flag as such:

    lclusterc --checkjob=uuid

Where uuid is the assigned number of the job in question.


To remove a job from the server, use the removejob argument:

    lclusterc --removejob=uuid

Where uuid is the assigned number of the job in question.


# TODOs

* merge in tests (only partly works)
* handling failed nodes
* watching warned jobs
* get nodes to send console output data back to server
* consider logging / recording past jobs


# Authors

This was created by Robert Bisewski at Ibis Cybernetics. For more
information, please contact us at:

* Website -> www.ibiscybernetics.com

* Email -> contact@ibiscybernetics.com
