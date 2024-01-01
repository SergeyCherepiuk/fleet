package node

import (
	"bufio"
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
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return errMemoryStat, err
	}
	defer file.Close()

	var stat MemoryStat
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

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
