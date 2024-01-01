package format

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type AccessMap[T any] map[string]func(T) any

func (am AccessMap[T]) HasAllHeaders(headers []string) bool {
	headersCopy := make([]string, len(headers))
	copy(headersCopy, headers)
	sort.Strings(headersCopy)

	accessMapKeys := maps.Keys(am)
	sort.Strings(accessMapKeys)

	return slices.Equal(headersCopy, accessMapKeys)
}

func Table[T any](headers []string, accessMap AccessMap[T], data []T) string {
	if len(headers) == 0 || !accessMap.HasAllHeaders(headers) {
		return ""
	}

	var (
		columns    = make([][]string, len(headers))
		maxLengths = make(map[string]int, len(headers))
	)

	for i := range headers {
		columns[i] = make([]string, 0, len(data)+1)
	}

	for i, header := range headers {
		columns[i] = append(columns[i], header)
		access := accessMap[header]
		for _, d := range data {
			value := fmt.Sprint(access(d))
			columns[i] = append(columns[i], value)
		}

		maxLengths[header] = maxLength(columns[i])

		maxLength := maxLengths[header]
		for j, value := range columns[i] {
			paddingSize := maxLength - len(value)
			columns[i][j] = value + strings.Repeat(" ", paddingSize)
		}
	}

	var builder strings.Builder

	for i := 0; i < len(data)+1; i++ {
		row := make([]string, 0, len(headers))
		for j := 0; j < len(headers); j++ {
			row = append(row, columns[j][i])
		}

		builder.WriteString(strings.Join(row, "  "))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func maxLength(elements []string) int {
	if len(elements) == 0 {
		return 0
	}

	max := len(elements[0])
	for i := 1; i < len(elements); i++ {
		l := len(elements[i])
		if l > max {
			max = l
		}
	}
	return max
}
