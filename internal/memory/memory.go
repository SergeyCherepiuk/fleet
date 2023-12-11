package memory

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ErrTotalMemoryLineNotFound error

const BytesInKilobyte = 1000

func Total() (uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.HasPrefix(line, "MemTotal") {
			return parseLine(line), nil
		}
	}

	return 0, ErrTotalMemoryLineNotFound(
		errors.New("total memory line not found in /proc/meminfo file"),
	)
}

func parseLine(line string) uint64 {
	pattern := regexp.MustCompile("[0-9]+")
	match := pattern.Find([]byte(line))
	if match == nil {
		return 0
	}

	value, err := strconv.ParseUint(string(match), 10, 64)
	if err != nil {
		return 0
	}

	return value * BytesInKilobyte
}
