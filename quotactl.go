package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	UsrQuota = iota
	GrpQuota
)

const (
	subCmdShift = 8
	subCmdMask  = 0x00ff
)

const (
	qGetQuota = 0x800007
)

func qCmd(subCmd, qType int) int {
	return subCmd<<subCmdShift | qType&subCmdMask
}

type Dqblk struct {
	DqbBHardlimit uint64
	DqbBSoftlimit uint64
	DqbCurSpace   uint64
	DqbIHardlimit uint64
	DqbISoftlimit uint64
	DqbCurInodes  uint64
	DqbBTime      uint64
	DqbITime      uint64
	DqbValid      uint32
}

func GetQuota(typ int, special string, id int) (result *Dqblk, err error) {
	result = &Dqblk{}
	if err = quotactl(qCmd(qGetQuota, typ), special, id, unsafe.Pointer(result)); err != nil {
		result = nil
	}
	return
}

func quotactl(cmd int, special string, id int, target unsafe.Pointer) (err error) {
	var deviceNamePtr *byte
	if deviceNamePtr, err = syscall.BytePtrFromString(special); err != nil {
		return
	}

	if _, _, errno := syscall.RawSyscall6(syscall.SYS_QUOTACTL, uintptr(cmd), uintptr(unsafe.Pointer(deviceNamePtr)), uintptr(id), uintptr(target), 0, 0); errno != 0 {
		err = os.NewSyscallError("quotactl", errno)
	}

	return
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s special uid\n", os.Args[0])
		os.Exit(1)
	}

	id, err := strconv.ParseUint(os.Args[2], 10, 32)
	if err != nil {
		fmt.Printf("Could not parse uid: %s\n", err.Error())
		os.Exit(2)
	}

	result, err := GetQuota(UsrQuota, os.Args[1], int(id))
	if err != nil {
		fmt.Printf("Could not retrieve quota: %s\n", err.Error())
		os.Exit(3)
	}

	fmt.Println("Space (1K Blocks):")
	fmt.Printf("  - hard limit: %d\n", result.DqbBHardlimit)
	fmt.Printf("  - soft limit: %d\n", result.DqbBSoftlimit)
	fmt.Printf("  - usage     : %d\n", result.DqbCurSpace)
	fmt.Println("Inodes:")
	fmt.Printf("  - hard limit: %d\n", result.DqbIHardlimit)
	fmt.Printf("  - soft limit: %d\n", result.DqbISoftlimit)
	fmt.Printf("  - usage     : %d\n", result.DqbCurInodes)
}
