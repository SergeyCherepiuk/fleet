package node

import (
	"bytes"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const BytesInKilobyte = 1000

type MemoryStat struct {
	Total     uint64
	Available uint64
}

var errMemoryStat = MemoryStat{Total: 0, Available: 0}

func Memory() (MemoryStat, error) {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return errMemoryStat, err
	}
	content = bytes.TrimSpace(content)

	var stat MemoryStat
	for _, line := range strings.Split(string(content), "\n") {
		switch {
		case strings.HasPrefix(line, "MemTotal"):
			stat.Total = parseMemoryLine(line)
		case strings.HasPrefix(line, "MemAvailable"):
			stat.Available = parseMemoryLine(line)
		}
	}

	return stat, nil
}

func parseMemoryLine(line string) uint64 {
	pattern := regexp.MustCompile("[0-9]+")
	match := pattern.Find([]byte(line))
	if match == nil {
		return 0
	}

	value, _ := strconv.ParseUint(string(match), 10, 64)
	return value * BytesInKilobyte
}
