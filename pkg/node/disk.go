package node

import (
	"bytes"
	"os"
	"syscall"
)

type DiskStat struct {
	Total     uint64
	Available uint64
}

var errDiskStat = DiskStat{Total: 0, Available: 0}

func Disk() (DiskStat, error) {
	mounts, err := mounts()
	if err != nil {
		return errDiskStat, err
	}

	var stat DiskStat
	for _, mount := range mounts {
		var buf syscall.Statfs_t
		if err := syscall.Statfs(mount, &buf); err != nil {
			continue
		}

		stat.Total += uint64(buf.Bsize) * buf.Blocks
		stat.Available += uint64(buf.Bsize) * buf.Bavail
	}
	return stat, nil
}

func mounts() ([]string, error) {
	content, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return make([]string, 0), err
	}
	content = bytes.TrimSpace(content)

	mounts := make([]string, 0)
	for _, line := range bytes.Split(content, []byte{'\n'}) {
		parts := bytes.Split(line, []byte{' '})
		if len(parts) < 2 {
			continue
		}
		mounts = append(mounts, string(parts[1]))
	}
	return mounts, nil
}
