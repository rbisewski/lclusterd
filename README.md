# lclusterd - a lite clustered job scheduler for Linux

lcluster is a lightweight Linux job scheduler designed to create clustered
nodes for the purposes of running small jobs in isolated libcontainer
instances.

Consider reading the *Basic Usage Instructions* at a minimum to get the hang
of using this job scheduler.

NOTE: this is *very* much a work in progress, and it does not currently
function.

# Requirements

The program itself was designed around a standard Debian Linux environment,
with the following requirements:

* linux kernel 3.19+
* golang 1.7+
* grpc

Popular distros, such as Debian or Fedora, will likely already contain
Linux kernel 4.0 or higher. Most or all of the above requirements are
probably going to satisfied given the modern nature of golang, which will
simply download the referenced repo as needed.

In the event that this program does not appear to work on a particular
non-mainstream distro. Feel free to contact me if you need assistance
and I will make note of it in future versions of this readme.


# Installation

You can obtain the codebase via:

go get github.com/rbisewski/lclusterd

Afterwards you can build both the server binary (lclusterd) and the client
binary (lclusterc) as follows:

make

If you need to make changes to the PB prototype file, you can also run the
following command:

make regen_proto

Note that prototypes are regenerated before every build using the standard
make target, so manual regeneration is not required.


# Basic Usage Instructions

To start the server daemon:

lclusterd [--namespace=name] [--rootfs=/path/to/dir]

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


# Authors

This was created by Robert Bisewski at Ibis Cybernetics. For more
information, please contact us at:

* Website -> www.ibiscybernetics.com

* Email -> contact@ibiscybernetics.com
