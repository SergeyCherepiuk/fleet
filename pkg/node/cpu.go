package node

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SergeyCherepiuk/fleet/internal/math"
	"golang.org/x/exp/maps"
)

type CPUStat struct {
	Cores uint
	Usage float64
}

var ErrNoCoresAvailable = errors.New("no cpu info available")

var errCPUStat = CPUStat{Cores: 0, Usage: 100.0}

func (*Node) CPU(interval time.Duration) (CPUStat, error) {
	stats := make([]map[string]uint64, 2)
	var cores uint

	for i := 0; i < 2; i++ {
		file, err := os.Open("/proc/stat")
		if err != nil {
			return errCPUStat, err
		}
		defer file.Close()

		lines := make([]string, 0)
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			if strings.HasPrefix(line, "cpu") {
				lines = append(lines, line)
			}
		}

		if len(lines) == 0 {
			return errCPUStat, ErrNoCoresAvailable
		}

		cores = uint(len(lines) - 1)
		stats[i] = parseCPULine(lines[0])

		if i == 0 {
			time.Sleep(interval)
		}
	}

	usage := usage(diff(stats[0], stats[1]))
	return CPUStat{Cores: cores, Usage: usage}, nil
}

func parseCPULine(line string) map[string]uint64 {
	columns := []string{
		"user", "nice", "system", "idle", "iowait",
		"irq", "softirq", "steal", "guest", "guest_nice",
	}
	stat := make(map[string]uint64)

	for strings.Contains(line, "  ") {
		line = strings.ReplaceAll(line, "  ", " ")
	}

	for i, s := range strings.Split(line, " ")[1:] {
		value, _ := strconv.ParseUint(s, 10, 64)
		stat[columns[i]] = value
	}

	return stat
}

func diff(stat1, stat2 map[string]uint64) map[string]uint64 {
	for k, v := range stat2 {
		stat2[k] = v - stat1[k]
	}
	return stat2
}

func usage(stat map[string]uint64) float64 {
	idleTime := float64(stat["idle"] + stat["iowait"])
	totalTime := float64(math.Sum(maps.Values(stat)))
	if totalTime == 0 {
		return 100.0
	}

	idlePercent := idleTime / totalTime * 100.0
	return 100.0 - idlePercent
}
