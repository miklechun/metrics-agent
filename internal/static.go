package internal

import (
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/jkstack/anet"
	"github.com/jkstack/jkframe/logging"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
)

func getStatic() *anet.HMStaticPayload {
	var ret anet.HMStaticPayload
	ret.Time = time.Now()
	fillStaticHostInfo(&ret)
	fillStaticCpuInfo(&ret)
	fillStaticMemoryInfo(&ret)
	fillStaticDiskInfo(&ret)
	fillStaticNetworkInfo(&ret)
	return &ret
}

func fillStaticHostInfo(ret *anet.HMStaticPayload) {
	info, err := host.Info()
	if err != nil {
		logging.Warning("get host.info: %v", err)
		return
	}
	ret.Host.Name = info.Hostname
	ret.Host.UpTime = time.Duration(info.Uptime) * time.Second
	ret.OS.Name = info.OS
	ret.OS.PlatformName = info.Platform
	ret.OS.PlatformVersion = info.PlatformVersion
	ret.OS.Install = time.Unix(0, 0) // TODO
	ret.Kernel.Version = info.KernelVersion
	ret.Kernel.Arch = info.KernelArch
}

func fillStaticCpuInfo(ret *anet.HMStaticPayload) {
	var err error
	ret.CPU.Physical, err = cpu.Counts(false)
	if err != nil {
		logging.Warning("get physical cpu count: %v", err)
	}
	ret.CPU.Logical, err = cpu.Counts(true)
	if err != nil {
		logging.Warning("get logical cpu count: %v", err)
	}
	cores, err := cpu.Info()
	if err != nil {
		logging.Warning("get cpu.info: %v", err)
		return
	}
	for _, core := range cores {
		id, _ := strconv.ParseUint(core.CoreID, 10, 32)
		physical, _ := strconv.ParseUint(core.PhysicalID, 10, 32)
		ret.CPU.Cores = append(ret.CPU.Cores, anet.HMCore{
			Processor: core.CPU,
			Model:     core.ModelName,
			Core:      int32(id),
			Cores:     core.Cores,
			Physical:  int32(physical),
			Mhz:       core.Mhz,
		})
	}
}

func fillStaticMemoryInfo(ret *anet.HMStaticPayload) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		logging.Warning("get memory.info: %v", err)
	}
	if vm != nil {
		ret.Memory.Physical = vm.Total
	}
	swap, err := mem.SwapMemory()
	if err != nil {
		logging.Warning("get swap.info: %v", err)
	}
	if swap != nil {
		ret.Memory.Swap = swap.Total
	}
}

func fillStaticDiskInfo(ret *anet.HMStaticPayload) {
	block, err := ghw.Block()
	if err != nil {
		logging.Warning("get blocks: %v", err)
	}
	for _, disk := range block.Disks {
		if disk.StorageController == ghw.STORAGE_CONTROLLER_UNKNOWN {
			continue
		}
		var parts []string
		for _, part := range disk.Partitions {
			if runtime.GOOS == "linux" {
				part.Name = "/dev/" + part.Name
			}
			parts = append(parts, part.Name)
		}
		ret.Disks = append(ret.Disks, anet.HMDisk{
			Model:      disk.Model,
			Total:      disk.SizeBytes,
			Type:       disk.DriveType.String(),
			Partitions: parts,
		})
	}
	parts, err := disk.Partitions(false)
	if err != nil {
		logging.Warning("get partitions: %v", err)
	}
	for _, part := range parts {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			logging.Warning("get partition usage(%s): %v", part.Mountpoint, err)
		}
		info := anet.HMPartition{
			Name:   part.Mountpoint,
			FSType: part.Fstype,
			Opts:   part.Opts,
		}
		if usage != nil {
			info.Total = usage.Total
		}
		ret.Partitions = append(ret.Partitions, info)
	}
}

func fillStaticNetworkInfo(ret *anet.HMStaticPayload) {
	ret.GateWay = "TODO"
	intfs, err := net.Interfaces()
	if err != nil {
		logging.Warning("get interfaces: %v", err)
	}
	for _, intf := range intfs {
		addrs, _ := intf.Addrs()
		ips := make([]string, len(addrs))
		for i, addr := range addrs {
			ips[i] = addr.String()
		}
		ret.Interface = append(ret.Interface, anet.HMInterface{
			Index:   intf.Index,
			Name:    intf.Name,
			Mtu:     intf.MTU,
			Flags:   strings.Split(intf.Flags.String(), "|"),
			Mac:     intf.HardwareAddr.String(),
			Address: ips,
		})
	}
}