package format

import (
	"fmt"
	"strings"
)

type AccessMap[T any] map[string]func(T) any

func Table[T any](headers []string, accessMap AccessMap[T], data []T) string {
	// TODO (SergeyCherepiuk): The check should be improved
	if len(headers) == 0 || len(headers) != len(accessMap) {
		return ""
	}

	var (
		values             = make(map[string][]string, len(headers))
		maxLengths         = make(map[string]int, len(headers))
		headersWithPadding = make([]string, 0, len(headers))
	)

	for _, header := range headers {
		columnValues := make([]string, 0, len(data))
		access := accessMap[header]
		for _, d := range data {
			value := fmt.Sprint(access(d))
			columnValues = append(columnValues, value)
		}
		values[header] = columnValues

		column := append([]string{header}, values[header]...)
		maxLengths[header] = maxLength(column)

		m := maxLengths[header]
		headerWithPadding := header + strings.Repeat(" ", m-len(header))
		headersWithPadding = append(headersWithPadding, headerWithPadding)
		for i, value := range values[header] {
			values[header][i] = value + strings.Repeat(" ", m-len(value))
		}
	}

	var builder strings.Builder

	builder.WriteString(strings.Join(headersWithPadding, "  "))
	builder.WriteByte('\n')

	for i := 0; i < len(values[headers[0]]); i++ {
		row := make([]string, 0, len(headers))
		for _, header := range headers {
			row = append(row, values[header][i])
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
