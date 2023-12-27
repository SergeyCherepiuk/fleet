package http

import (
	"encoding/json"
	"io"
)

func ErrorMessage(body io.ReadCloser) string {
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return ""
	}

	var b struct{ Message string }
	if err := json.Unmarshal(data, &b); err != nil {
		return ""
	}

	return b.Message
}
