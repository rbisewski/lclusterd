/*
 * File: libcontainer_funcs.go
 *
 * Description: functions to assist container spawning
 */

package libetcd

import (
	"../../lcfg"
	"github.com/opencontainers/runc/libcontainer/configs"
	"syscall"
)

//! Assemble and return a libcontainer config.
/*
 * @param    string            pseudo-random name of container
 * @param    string            given hostname
 * @param    string            rootfs
 *
 * @return   configs.Config    libcontainer config
 */
func GenerateRuncConfig(containerName string, hostname string, rootfs string) *configs.Config {

	// Set the default mounting flags.
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// Assemble the new config for the container.
	newConfig := &configs.Config{

		Rootfs: rootfs,

		Capabilities: &configs.Capabilities{
			Bounding:    lcfg.LclustercCaps,
			Permitted:   lcfg.LclustercCaps,
			Inheritable: lcfg.LclustercCaps,
			Ambient:     lcfg.LclustercCaps,
			Effective:   lcfg.LclustercCaps,
		},

		Namespaces: configs.Namespaces([]configs.Namespace{
			{Type: configs.NEWNS},
			{Type: configs.NEWUTS},
			{Type: configs.NEWIPC},
			{Type: configs.NEWPID},
			{Type: configs.NEWNET},
		}),

		Cgroups: &configs.Cgroup{
			Name:   containerName,
			Parent: "system",
			Resources: &configs.Resources{
				MemorySwappiness: nil,
				AllowedDevices:   configs.DefaultAllowedDevices,
			},
		},

		MaskPaths: []string{
			"/proc/kcore",
		},

		ReadonlyPaths: []string{
			"/proc/sys", "/proc/sysrq-trigger", "/proc/irq", "/proc/bus",
		},

		Devices: configs.DefaultAutoCreatedDevices,

		Hostname: hostname,

		Mounts: []*configs.Mount{
			{
				Source:      "proc",
				Destination: "/proc",
				Device:      "proc",
				Flags:       defaultMountFlags,
			},
			{
				Source:      "tmpfs",
				Destination: "/dev",
				Device:      "tmpfs",
				Flags:       syscall.MS_NOSUID | syscall.MS_STRICTATIME,
				Data:        "mode=755",
			},
			{
				Source:      "devpts",
				Destination: "/dev/pts",
				Device:      "devpts",
				Flags:       syscall.MS_NOSUID | syscall.MS_NOEXEC,
				Data:        "newinstance,ptmxmode=0666,mode=0620,gid=5",
			},
			{
				Device:      "tmpfs",
				Source:      "shm",
				Destination: "/dev/shm",
				Data:        "mode=1777,size=65536k",
				Flags:       defaultMountFlags,
			},
			{
				Source:      "mqueue",
				Destination: "/dev/mqueue",
				Device:      "mqueue",
				Flags:       defaultMountFlags,
			},
			{
				Source:      "sysfs",
				Destination: "/sys",
				Device:      "sysfs",
				Flags:       defaultMountFlags | syscall.MS_RDONLY,
			},
		},

		Networks: []*configs.Network{
			{
				Type:    "loopback",
				Address: "127.0.0.1/0",
				Gateway: "localhost",
			},
		},

		Rlimits: []configs.Rlimit{
			{
				Type: syscall.RLIMIT_NOFILE,
				Hard: uint64(1025),
				Soft: uint64(1025),
			},
		},
	}

	// Go ahead and return the newly generated config
	return newConfig
}
